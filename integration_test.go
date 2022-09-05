//go:build integration
// +build integration

package runn

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

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
		t.Run(tt.book, func(t *testing.T) {
			ctx := context.Background()
			f, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			if err := f.Run(ctx); err != nil {
				t.Error(err)
			}
		})
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
		t.Run(tt.book, func(t *testing.T) {
			ctx := context.Background()
			f, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			if err := f.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestRunUsingMySQL(t *testing.T) {
	db := createMySQLContainer(t)
	tests := []struct {
		book string
	}{
		// TODO: Add runbook
		// {"testdata/book/mysql.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			ctx := context.Background()
			f, err := New(Book(tt.book), DBRunner("db", db))
			if err != nil {
				t.Fatal(err)
			}
			if err := f.Run(ctx); err != nil {
				t.Error(err)
			}
		})
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

func createMySQLContainer(t *testing.T) *sql.DB {
	t.Helper()
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	opt := &dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "8",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=rootpass",
			"MYSQL_USER=myuser",
			"MYSQL_PASSWORD=mypass",
			"MYSQL_DATABASE=testdb",
		},
		Mounts: []string{
			fmt.Sprintf("%s:/docker-entrypoint-initdb.d/initdb.sql", filepath.Join(wd, "testdata", "initdb", "mysql.sql")),
		},
		Cmd: []string{
			"mysqld",
			"--character-set-server=utf8mb4",
			"--collation-server=utf8mb4_unicode_ci",
		},
	}
	my, err := pool.RunWithOptions(opt)
	if err != nil {
		t.Fatalf("Could not start resource: %s", err)
	}

	var db *sql.DB
	if err := pool.Retry(func() error {
		time.Sleep(time.Second * 10)
		var err error
		db, err = sql.Open("mysql", fmt.Sprintf("myuser:mypass@(localhost:%s)/testdb?parseTime=true", my.GetPort("3306/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		t.Fatalf("Could not connect to database: %s", err)
	}

	t.Cleanup(func() {
		if err := pool.Purge(my); err != nil {
			t.Fatalf("Could not purge resource: %s", err)
		}
	})

	return db
}
