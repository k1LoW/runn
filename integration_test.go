//go:build integration

package runn

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/k1LoW/runn/testutil"
	"github.com/xo/dburl"
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
		{"testdata/book/cookie_in_requests_automatically.yml"},
		{"testdata/book/cookie.yml"},
		{"testdata/book/http_with_use_trace.yml"},
		{"testdata/book/http_with_use_trace_header_name.yml"},
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
	db, _ := testutil.CreateMySQLContainer(t)
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

func TestMultipleConnections(t *testing.T) {
	db, dsn := testutil.CreateMySQLContainer(t)
	gs := testutil.GRPCServer(t, false, false)
	dir := t.TempDir()
	book := "testdata/book/multiple_conn.yml"
	b, err := os.ReadFile(book)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 256; i++ {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d.yml", i)), b, os.ModePerm); err != nil {
			t.Fatal(err)
		}
	}
	ctx := context.Background()
	t.Setenv("TEST_DB", dsn)
	t.Setenv("TEST_GRPC", gs.Addr())
	opts := []Option{
		T(t), FailFast(true),
		DBRunner("db", db),
		GrpcRunner("greq", gs.ClientConn()),
		Scopes(ScopeAllowReadParent),
	}
	o, err := Load(filepath.Join(dir, "*.yml"), opts...)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.RunN(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestRunUsingSSHd(t *testing.T) {
	_, host, hostname, user, port := testutil.CreateSSHdContainer(t)
	t.Setenv("TEST_HOST", host)
	t.Setenv("TEST_HOSTNAME", hostname)
	t.Setenv("TEST_USER", user)
	t.Setenv("TEST_PORT", strconv.Itoa(port))
	b, err := os.ReadFile(filepath.Join("testdata", "sshd", "id_rsa"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TEST_PRIVATE_KEY", string(b))
	tests := []struct {
		book string
	}{
		{"testdata/book/sshd.yml"},
		{"testdata/book/sshd_no_config.yml"},
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

func TestSSHPortFowarding(t *testing.T) {
	_ = testutil.CreateHTTPBinContainer(t)
	_, host, hostname, user, port := testutil.CreateSSHdContainer(t)
	_, _ = testutil.CreateMySQLContainer(t)
	t.Setenv("TEST_HOST", host)
	t.Setenv("TEST_HOSTNAME", hostname)
	t.Setenv("TEST_USER", user)
	t.Setenv("TEST_PORT", strconv.Itoa(port))
	t.Setenv("TEST_HTTP_FOWARD_PORT", strconv.Itoa(testutil.NewPort(t)))
	t.Setenv("TEST_DB_FOWARD_PORT", strconv.Itoa(testutil.NewPort(t)))
	tests := []struct {
		book string
	}{
		{"testdata/book/sshd_local_forward.yml"},
		{"testdata/book/sshd_local_forward_with_openapi3.yml"},
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

func TestRunViaHTTPS(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/http.yml"},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/http_multipart.yml"},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/grpc.yml"},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/db.yml"},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/exec.yml"},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/include_main.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			hs := testutil.HTTPServer(t)
			gs := testutil.GRPCServer(t, false, false)
			db, _ := testutil.SQLite(t)
			opts := []Option{
				Book(tt.book),
				HTTPRunner("req", hs.URL, hs.Client(), MultipartBoundary(testutil.MultipartBoundary)),
				GrpcRunner("greq", gs.Conn()),
				DBRunner("db", db),
				Func("upcase", strings.ToUpper),
				Scopes(ScopeAllowReadRemote, ScopeAllowRunExec),
			}
			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestRunViaGitHub(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"github://k1LoW/runn/testdata/book/http.yml"},
		{"github://k1LoW/runn/testdata/book/http_multipart.yml"},
		{"github://k1LoW/runn/testdata/book/grpc.yml"},
		{"github://k1LoW/runn/testdata/book/db.yml"},
		{"github://k1LoW/runn/testdata/book/exec.yml"},
		{"github://k1LoW/runn/testdata/book/include_main.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			hs := testutil.HTTPServer(t)
			gs := testutil.GRPCServer(t, false, false)
			db, _ := testutil.SQLite(t)
			opts := []Option{
				Book(tt.book),
				HTTPRunner("req", hs.URL, hs.Client(), MultipartBoundary(testutil.MultipartBoundary)),
				GrpcRunner("greq", gs.Conn()),
				DBRunner("db", db),
				Func("upcase", strings.ToUpper),
				Scopes(ScopeAllowReadRemote, ScopeAllowRunExec),
			}
			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestRunUsingHTTPBinTimeout(t *testing.T) {
	host := testutil.CreateHTTPBinContainer(t)
	t.Setenv("HTTPBIN_END_POINT", host)
	tests := []struct {
		book string
	}{
		{"testdata/book/http_timeout.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			ctx := context.Background()
			o, err := New(Book(tt.book), Stdout(io.Discard), Stderr(io.Discard))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err == nil {
				t.Errorf("No timeout error occurred.")
			}
		})
	}
}

func TestHostRulesWithContainer(t *testing.T) {
	ctx := context.Background()
	t.Run("DB", func(t *testing.T) {
		tests := []struct {
			book string
		}{
			{"testdata/book/db_with_host_rules.yml"},
			{"testdata/book/db_with_host_rules_wildcard.yml"},
		}
		_, dsn := testutil.CreateMySQLContainer(t)
		u, err := dburl.Parse(dsn)
		if err != nil {
			t.Fatal(err)
		}
		t.Setenv("TEST_DB_HOST_RULE", u.Host)
		for _, tt := range tests {
			t.Run(tt.book, func(t *testing.T) {
				o, err := New(Book(tt.book))
				if err != nil {
					t.Fatal(err)
					return
				}
				if err := o.Run(ctx); err != nil {
					t.Error(err)
				}
			})
		}
	})
}
