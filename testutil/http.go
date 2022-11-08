package testutil

import (
	"net/http"
	"net/http/httptest"
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
    <input name="upload" type="file"/>
    <input name="submit" type="submit"/>
  </form>
</body>
</html>
`

func HTTPServer(t *testing.T) *httptest.Server {
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
	r.Method(http.MethodGet).Path("/hello").Header("Content-Type", "text/html; charset=utf-8").ResponseString(http.StatusOK, "<h1>Hello</h1>")
	r.Method(http.MethodPost).Path("/upload").Header("Content-Type", "text/html; charset=utf-8").ResponseString(http.StatusOK, "<h1>Posted</h1>")
	r.Method(http.MethodGet).Header("Content-Type", "text/html; charset=utf-8").ResponseString(http.StatusNotFound, "<h1>Not Found</h1>")
	ts := r.Server()
	t.Cleanup(func() {
		ts.Close()
	})

	return ts
}
