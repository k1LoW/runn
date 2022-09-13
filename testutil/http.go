package testutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/k1LoW/httpstub"
)

func HTTPServer(t *testing.T) *httptest.Server {
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
