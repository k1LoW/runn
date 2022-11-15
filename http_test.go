package runn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/runn/testutil"
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
		r, err := newHTTPRunner("req", endpoint)
		if err != nil {
			t.Fatal(err)
		}
		r.operator = o
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
		res, ok := r.operator.store.steps[i]["res"].(map[string]interface{})
		if !ok {
			t.Fatalf("invalid steps res: %v", r.operator.store.steps[i]["res"])
		}
		got, ok := res["status"].(int)
		if !ok {
			t.Fatalf("invalid res status: %v", res["status"])
		}
		if got != tt.want {
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

	multitests := []struct {
		in           string
		mediaType    string
		wantContains []string
	}{
		{
			`
file1: 'testdata/dummy.png'
file2: 'testdata/dummy.jpeg'`,
			MediaTypeMultipartFormData,
			[]string{},
		},
	}

	for _, tt := range multitests {
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
		for _, want := range tt.wantContains {
			if !strings.Contains(got, want) {
				t.Errorf("got %v\nexpect to contain %v", got, want)
			}
		}
		contentType := r.multipartWriter.FormDataContentType()
		if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
			t.Errorf("got %v\nexpect to has prefix %v", contentType, "multipart/form-data; boundary=")
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
		r, err := newHTTPRunnerWithHandler(t.Name(), s)
		if err != nil {
			t.Fatal(err)
		}
		r.operator = o
		if err := r.Run(ctx, tt.req); err != nil {
			t.Error(err)
			continue
		}
		res, ok := r.operator.store.steps[i]["res"].(map[string]interface{})
		if !ok {
			t.Fatalf("invalid steps res: %v", r.operator.store.steps[i]["res"])
		}
		got, ok := res["status"].(int)
		if !ok {
			t.Fatalf("invalid res status: %v", res["status"])
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestNotFollowRedirect(t *testing.T) {
	tests := []struct {
		req               *httpRequest
		notFollowRedirect bool
		want              int
	}{
		{
			&httpRequest{
				path:    "/redirect",
				method:  http.MethodGet,
				headers: map[string]string{},
			},
			false,
			http.StatusNotFound,
		},
		{
			&httpRequest{
				path:    "/redirect",
				method:  http.MethodGet,
				headers: map[string]string{},
			},
			true,
			http.StatusFound,
		},
	}
	ctx := context.Background()
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	hs := testutil.HTTPServer(t)
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.req), func(t *testing.T) {
			r, err := newHTTPRunner("req", hs.URL)
			if err != nil {
				t.Fatal(err)
			}
			r.operator = o
			if tt.notFollowRedirect {
				r.client.CheckRedirect = notFollowRedirectFn
			}
			if err := r.Run(ctx, tt.req); err != nil {
				t.Error(err)
				return
			}
			res, ok := r.operator.store.latest()["res"].(map[string]interface{})
			if !ok {
				t.Fatalf("invalid res: %#v", r.operator.store.latest()["res"])
			}
			got, ok := res["status"].(int)
			if !ok {
				t.Fatalf("invalid res status: %v", res["status"])
			}
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}
