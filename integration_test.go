////go:build integration

package runn

import (
	"context"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/k1LoW/runn/testutil"
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
	host := testutil.CreateHTTPBinContainer(t)
	t.Setenv("HTTPBIN_END_POINT", host)
	tests := []struct {
		book string
	}{
		{"testdata/book/httpbin.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			ctx := context.Background()
			o, err := New(Book(tt.book), Stdout(io.Discard), Stderr(io.Discard))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestRunUsingMySQL(t *testing.T) {
	db := testutil.CreateMySQLContainer(t)
	tests := []struct {
		book string
	}{
		{"testdata/book/mysql.yml"},
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

func TestRunUsingSSHd(t *testing.T) {
	_, host, hostname, user, port := testutil.CreateSSHdContainer(t)
	t.Setenv("TEST_HOST", host)
	t.Setenv("TEST_HOSTNAME", hostname)
	t.Setenv("TEST_USER", user)
	t.Setenv("TEST_PORT", strconv.Itoa(port))
	tests := []struct {
		book string
	}{
		{"testdata/book/sshd.yml"},
		{"testdata/book/sshd_keep_session.yml"},
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

func TestUsingPkgGoDev(t *testing.T) {
	if testutil.SkipCDPTest(t) {
		t.Skip("chrome not found")
	}
	tests := []struct {
		book string
	}{
		{"testdata/book/pkg_go_dev.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.book, func(t *testing.T) {
			t.Parallel()
			o, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}
