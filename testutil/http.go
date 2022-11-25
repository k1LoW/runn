package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/k1LoW/httpstub"
)

const formHTML = `<!doctype html>
<html>
<head>
  <title>For runn test</title>
</head>
<body>
  <header>
    <h1 class="runn-test" data-test-id="runn-h1">Test Form</h1>
    <a href="/hello">Link</a>
  </header>
  <form class="form-test" method="POST" action="/upload" enctype="multipart/form-data">
    <input name="username" type="text"/>
    <input name="upload0" type="file"/>
    <input name="upload1" type="file"/>
    <input name="submit" type="submit"/>
  </form>
  <script>
	localStorage.setItem('local', 'storage');
	sessionStorage.setItem('session', 'storage');
  </script>
</body>
</html>
`
const MultipartBoundary = "123456789012345678901234567890abcdefghijklmnopqrstuvwxyz"

func HTTPServer(t *testing.T) *httptest.Server {
	ts, _ := HTTPServerAndRouter(t)
	return ts
}

func HTTPServerAndRouter(t *testing.T) (*httptest.Server, *httpstub.Router) {
	r := httpstub.NewRouter(t)
	r.Method(http.MethodPost).Path("/users").Response(http.StatusCreated, nil)
	r.Method(http.MethodPost).Path("/help").Response(http.StatusCreated, nil)
	r.Method(http.MethodGet).Path("/users/1").Header("Content-Type", "application/json").ResponseString(http.StatusOK, `{"data":{"username":"alice"}}`)
	r.Method(http.MethodGet).Path("/private").Match(func(r *http.Request) bool {
		ah := r.Header.Get("Authorization")
		return !strings.Contains(ah, "Bearer")
	}).Header("Content-Type", "application/json").ResponseString(http.StatusForbidden, `{"error":"Forbidden"}`)
	r.Method(http.MethodGet).Path("/private").Match(func(r *http.Request) bool {
		ah := r.Header.Get("Authorization")
		return strings.Contains(ah, "Bearer")
	}).Response(http.StatusOK, nil)
	r.Method(http.MethodGet).Path("/redirect").Header("Location", "/notfound").Response(http.StatusFound, nil)
	r.Method(http.MethodGet).Path("/form").Header("Content-Type", "text/html; charset=utf-8").ResponseString(http.StatusOK, formHTML)
	r.Method(http.MethodGet).Match(func(r *http.Request) bool {
		return strings.HasPrefix(r.URL.Path, "/increment/")
	}).Header("Content-Type", "application/json").Handler(func(w http.ResponseWriter, r *http.Request) {
		i, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/increment/"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"value": -1}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf(`{"value": %d}`, i+1)))
		return
	})
	r.Method(http.MethodGet).Path("/hello").Header("Content-Type", "text/html; charset=utf-8").ResponseString(http.StatusOK, "<h1>Hello</h1>")
	r.Method(http.MethodPost).Path("/upload").Header("Content-Type", "text/html; charset=utf-8").ResponseString(http.StatusCreated, "<h1>Posted</h1>")
	r.Method(http.MethodGet).Header("Content-Type", "text/html; charset=utf-8").ResponseString(http.StatusNotFound, "<h1>Not Found</h1>")
	ts := r.Server()
	t.Cleanup(func() {
		ts.Close()
	})

	return ts, r
}
