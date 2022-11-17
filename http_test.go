package runn

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
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
}

func TestRequestBodyForMultipart(t *testing.T) {
	t.Setenv("TEST_MODE", "true")
	dummy1, err := os.ReadFile("testdata/dummy.png")
	if err != nil {
		t.Fatal(err)
	}
	dummy2, err := os.ReadFile("testdata/dummy.jpeg")
	if err != nil {
		t.Fatal(err)
	}

	multitests := []struct {
		in              string
		mediaType       string
		wantBody        string
		wantContentType string
	}{
		{
			`
upload0: 'testdata/dummy.png'
upload1: 'testdata/dummy.jpeg'
name: 'bob'`,
			MediaTypeMultipartFormData,
			"--123456789012345678901234567890abcdefghijklmnopqrstuvwxyz\r\n" +
				strings.Join([]string{
					"Content-Disposition: form-data; name=\"upload0\"; filename=\"dummy.png\"\r\nContent-Type: image/png\r\n\r\n" + string(dummy1),
					"Content-Disposition: form-data; name=\"upload1\"; filename=\"dummy.jpeg\"\r\nContent-Type: image/jpeg\r\n\r\n" + string(dummy2),
					"Content-Disposition: form-data; name=\"name\"\r\n\r\nbob",
				}, "\r\n--123456789012345678901234567890abcdefghijklmnopqrstuvwxyz\r\n") +
				"\r\n--123456789012345678901234567890abcdefghijklmnopqrstuvwxyz--\r\n",
			"multipart/form-data; boundary=123456789012345678901234567890abcdefghijklmnopqrstuvwxyz",
		},
	}

	for idx, tt := range multitests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			var b interface{}
			if err := yaml.Unmarshal([]byte(tt.in), &b); err != nil {
				t.Error(err)
				return
			}
			r := &httpRequest{
				mediaType: tt.mediaType,
				body:      b,
			}
			body, err := r.encodeBody()
			if err != nil {
				t.Error(err)
				return
			}
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, body); err != nil {
				t.Error(err)
				return
			}
			got := buf.String()
			if diff := cmp.Diff(got, tt.wantBody, nil); diff != "" {
				t.Errorf("%s", diff)
			}
			contentType := r.multipartWriter.FormDataContentType()
			if contentType != tt.wantContentType {
				t.Errorf("got %v\nwant %v", got, tt.wantContentType)
			}
		})
	}
}

func TestRequestBodyForMultipart_onServer(t *testing.T) {
	t.Setenv("TEST_MODE", "true")
	dummy1, err := os.ReadFile("testdata/dummy.png")
	if err != nil {
		t.Fatal(err)
	}
	dummy2, err := os.ReadFile("testdata/dummy.jpeg")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		in                     string
		req                    *httpRequest
		wantContainRequestBody []string
	}{
		{
			`
upload0: 'testdata/dummy.png'
upload1: 'testdata/dummy.jpeg'
username: 'bob'`,
			&httpRequest{
				path:      "/upload",
				method:    http.MethodPost,
				mediaType: MediaTypeMultipartFormData,
			},
			[]string{
				"Content-Disposition: form-data; name=\"upload0\"; filename=\"dummy.png\"\r\nContent-Type: image/png\r\n\r\n" + string(dummy1),
				"Content-Disposition: form-data; name=\"upload1\"; filename=\"dummy.jpeg\"\r\nContent-Type: image/jpeg\r\n\r\n" + string(dummy2),
				"Content-Disposition: form-data; name=\"username\"\r\n\r\nbob",
			},
		},
	}

	ctx := context.Background()
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	hs, hr := testutil.HTTPServerAndRouter(t)
	for idx, tt := range tests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			var b interface{}
			if err := yaml.Unmarshal([]byte(tt.in), &b); err != nil {
				t.Error(err)
				return
			}
			tt.req.body = b
			r, err := newHTTPRunner("req", hs.URL)
			if err != nil {
				t.Error(err)
				return
			}
			r.operator = o
			if err := r.Run(ctx, tt.req); err != nil {
				t.Error(err)
				return
			}
			gotBody, err := io.ReadAll(hr.Requests()[0].Body)
			if err != nil {
				t.Error(err)
				return
			}
			for _, wb := range tt.wantContainRequestBody {
				if !strings.Contains(string(gotBody), wb) {
					t.Errorf("got %v\nwant to contain %v", string(gotBody), wb)
				}
			}
		})
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
