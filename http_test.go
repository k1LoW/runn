package runn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/httpstub"
)

func TestHTTPRunnerRunUsingGitHubAPI(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("env GITHUB_TOKEN is not set")
	}
	endpoint := "https://api.github.com"
	tests := []struct {
		req                  *httpRequest
		useOpenApi3Validator bool
		want                 int
	}{
		{
			&httpRequest{
				path:      "/users/k1LoW",
				method:    http.MethodGet,
				mediaType: MediaTypeApplicationJSON,
				headers: map[string]string{
					"Authorization": fmt.Sprintf("token %s", os.Getenv("GITHUB_TOKEN")),
				},
			},
			true,
			http.StatusOK,
		},
		{
			&httpRequest{
				path:      "/invalid/endpoint",
				method:    http.MethodGet,
				mediaType: MediaTypeApplicationJSON,
				headers:   map[string]string{},
			},
			false,
			http.StatusNotFound,
		},
	}

	ctx := context.Background()
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for i, tt := range tests {
		r, err := newHTTPRunner("req", endpoint, o)
		if err != nil {
			t.Fatal(err)
		}
		if tt.useOpenApi3Validator {
			c := &httpRunnerConfig{
				OpenApi3DocLocation:  "testdata/openapi3.yml",
				SkipValidateRequest:  false,
				SkipValidateResponse: false,
			}
			v, err := newHttpValidator(c)
			if err != nil {
				t.Fatal(err)
			}
			r.validator = v
		}
		if err := r.Run(ctx, tt.req); err != nil {
			t.Error(err)
			continue
		}
		if want := i + 1; len(r.operator.store.steps) != want {
			t.Errorf("got %v want %v", len(r.operator.store.steps), want)
			continue
		}
		res := r.operator.store.steps[i]["res"].(map[string]interface{})
		if got := res["status"].(int); got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestRequestBody(t *testing.T) {
	tests := []struct {
		in        string
		mediaType string
		want      string
	}{
		{
			`
data:
  one: ichi
  two: ni`,
			MediaTypeApplicationJSON,
			`{"data":{"one":"ichi","two":"ni"}}`,
		},
		{
			`
data:
  one: 1
  two: ni`,
			MediaTypeApplicationJSON,
			`{"data":{"one":1,"two":"ni"}}`,
		},
		{
			`text`,
			MediaTypeTextPlain,
			`text`,
		},
		{
			`
one: ichi
two: ni`,
			MediaTypeApplicationFormUrlencoded,
			`one=ichi&two=ni`,
		},
	}

	for _, tt := range tests {
		var b interface{}
		if err := yaml.Unmarshal([]byte(tt.in), &b); err != nil {
			t.Fatal(err)
		}
		r := &httpRequest{
			mediaType: tt.mediaType,
			body:      b,
		}
		body, err := r.encodeBody()
		if err != nil {
			t.Fatal(err)
		}
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, body); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestMergeURL(t *testing.T) {
	tests := []struct {
		endpoint string
		path     string
		want     string
	}{
		{"https://git.example.com/api/v3", "/orgs/octokit/repos", "https://git.example.com/api/v3/orgs/octokit/repos"},
		{"https://git.example.com/api/v3", "/repos/vmg/redcarpet/issues?state=closed", "https://git.example.com/api/v3/repos/vmg/redcarpet/issues?state=closed"},
	}
	for _, tt := range tests {
		u, err := url.Parse(tt.endpoint)
		if err != nil {
			t.Fatal(err)
		}
		got, err := mergeURL(u, tt.path)
		if err != nil {
			t.Error(err)
			continue
		}
		if got.String() != tt.want {
			t.Errorf("got %v\nwant %v", got.String(), tt.want)
		}
	}
}

func TestHTTPRunnerWithHandler(t *testing.T) {
	tests := []struct {
		req         *httpRequest
		pattern     string
		handlerFunc func(w http.ResponseWriter, r *http.Request)
		want        int
	}{
		{
			&httpRequest{
				path:      "/users/k1LoW",
				method:    http.MethodGet,
				mediaType: MediaTypeApplicationJSON,
			},
			"/users/k1LoW",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("hello k1LoW!"))
			},
			http.StatusOK,
		},
		{
			&httpRequest{
				path:      "/users/k1LoW",
				method:    http.MethodGet,
				mediaType: MediaTypeApplicationJSON,
			},
			"/users/unknownuser",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("hello k1LoW!"))
			},
			http.StatusNotFound,
		},
	}
	ctx := context.Background()
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for i, tt := range tests {
		s := http.NewServeMux()
		s.HandleFunc(tt.pattern, tt.handlerFunc)
		r, err := newHTTPRunnerWithHandler(t.Name(), s, o)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Run(ctx, tt.req); err != nil {
			t.Error(err)
			continue
		}
		res := r.operator.store.steps[i]["res"].(map[string]interface{})
		if got := res["status"].(int); got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func testHTTPServer(t *testing.T) *httptest.Server {
	r := httpstub.NewRouter(t)
	r.Method(http.MethodPost).Path("/users").Response(http.StatusCreated, nil)
	r.Method(http.MethodGet).Path("/users/1").Header("Content-Type", "application/json").ResponseString(http.StatusOK, `{"data":{"username":"alice"}}`)
	r.Method(http.MethodGet).Path("/private").Match(func(r *http.Request) bool {
		ah := r.Header.Get("Authorization")
		return !strings.Contains(ah, "Bearer")
	}).Header("Content-Type", "application/json").ResponseString(http.StatusForbidden, `{"error":"Forbidden"}`)
	r.Method(http.MethodGet).Path("/private").Match(func(r *http.Request) bool {
		ah := r.Header.Get("Authorization")
		return strings.Contains(ah, "Bearer")
	}).Response(http.StatusOK, nil)
	ts := r.Server()
	t.Cleanup(func() {
		ts.Close()
	})

	return ts
}
