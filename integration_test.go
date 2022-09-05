//go:build integration
// +build integration

package runn

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
)

func TestRunUsingGitHubAPI(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("env GITHUB_TOKEN is not set")
	}
	tests := []struct {
		book string
	}{
		{"testdata/book/github.yml"},
		{"testdata/book/github_map.yml"},
	}
	for _, tt := range tests {
		ctx := context.Background()
		f, err := New(Book(tt.book))
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestRunUsingHTTPBin(t *testing.T) {
	host := createHTTPBinContainer(t)
	t.Setenv("HTTPBIN_END_POINT", host)
	tests := []struct {
		book string
	}{
		{"testdata/book/httpbin.yml"},
	}
	for _, tt := range tests {
		ctx := context.Background()
		f, err := New(Book(tt.book))
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func createHTTPBinContainer(t *testing.T) string {
	t.Helper()
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	httpbin, err := pool.Run("kennethreitz/httpbin", "latest", []string{})
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	var host string
	if err := pool.Retry(func() error {
		host = fmt.Sprintf("http://localhost:%s/", httpbin.GetPort("80/tcp"))
		_, err := http.Get(host)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatalf("Could not connect to database: %s", err)
	}

	t.Cleanup(func() {
		if err := pool.Purge(httpbin); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	})
	return host
}
