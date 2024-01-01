package testutil

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// BenchmarkSet returns a server and a path pattern for benchmarking.
func BenchmarkSet(t testing.TB, bookCount, stepCount int, bodySize int) (*httptest.Server, string) {
	const httpRunbookTmpl = `desc: Test using HTTP
runners:
  req: {{ .Endpoint }}
steps:
{{ range $k := .Steps }}
-
  req:
    /:
      post:
        body:
          application/json:
            data: {{ $.Body }}
{{ end }}
`
	ts := echoServer(t)
	t.Cleanup(func() {
		ts.Close()
	})
	dir := t.TempDir()
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	})
	tmpl, err := template.New("http").Parse(httpRunbookTmpl)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < bookCount; i++ {
		book := filepath.Join(dir, fmt.Sprintf("http-%d.yml", i))
		f, err := os.Create(book)
		if err != nil {
			t.Fatal(err)
		}
		if err := tmpl.Execute(f, map[string]any{
			"Endpoint": ts.URL,
			"Steps":    make([]struct{}, stepCount),
			"Body":     strings.Repeat("a", bodySize),
		}); err != nil {
			t.Fatal(err)
		}
	}
	return ts, fmt.Sprintf("%s/*.yml", dir)
}

// echoServer returns a server that does not do anything special, just echo.
// For benchmarking.
func echoServer(t testing.TB) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, r.Body)
	}))
	t.Cleanup(func() {
		ts.Close()
	})
	return ts
}
