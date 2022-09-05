//go:build integration
// +build integration

package runn

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
)

var httpbin *dockertest.Resource

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	httpbin, err = pool.Run("kennethreitz/httpbin", "latest", []string{})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		_, err := http.Get(fmt.Sprintf("http://localhost:%s/", httpbin.GetPort("80/tcp")))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(httpbin); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

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

func TestRunUsingHttpbin(t *testing.T) {
	t.Setenv("HTTPBIN_END_POINT", fmt.Sprintf("http://localhost:%s/", httpbin.GetPort("80/tcp")))
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
