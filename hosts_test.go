package runn

import (
	"context"
	"strings"
	"testing"

	"github.com/k1LoW/runn/testutil"
)

func TestHostRules(t *testing.T) {
	ctx := context.Background()
	t.Run("HTTP", func(t *testing.T) {
		tests := []struct {
			book string
		}{
			{"testdata/book/http_with_host_rules.yml"},
			{"testdata/book/http_with_host_rules_wildcard.yml"},
		}
		ts, tr := testutil.HTTPServerAndRouter(t)
		t.Setenv("TEST_HTTP_HOST_RULE", strings.TrimPrefix(ts.URL, "http://"))
		for _, tt := range tests {
			t.Run(tt.book, func(t *testing.T) {
				tr.ClearRequests()
				o, err := New(Book(tt.book))
				if err != nil {
					t.Fatal(err)
					return
				}
				if err := o.Run(ctx); err != nil {
					t.Error(err)
					return
				}
				want := "app.example.com"
				if tr.Requests()[0].Host != "app.example.com" {
					t.Errorf("got %s want %s", tr.Requests()[0].Host, want)
				}
			})
		}
	})

	t.Run("gRPC", func(t *testing.T) {
		tests := []struct {
			book string
		}{
			{"testdata/book/grpc_with_host_rules.yml"},
			{"testdata/book/grpc_with_host_rules_wildcard.yml"},
		}
		ts := testutil.GRPCServer(t, true, false)
		t.Setenv("TEST_GRPC_HOST_RULE", ts.Addr())
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

	t.Run("CDP", func(t *testing.T) {
		tests := []struct {
			book string
		}{
			{"testdata/book/cdp_with_host_rules.yml"},
			{"testdata/book/cdp_with_host_rules_wildcard.yml"},
		}
		ts, tr := testutil.HTTPServerAndRouter(t)
		t.Setenv("TEST_HTTP_HOST_RULE", strings.TrimPrefix(ts.URL, "http://"))
		for _, tt := range tests {
			t.Run(tt.book, func(t *testing.T) {
				o, err := New(Book(tt.book))
				if err != nil {
					t.Fatal(err)
					return
				}
				if err := o.Run(ctx); err != nil {
					t.Error(err)
					return
				}
				want := "blog.example.com"
				if tr.Requests()[0].Host != "blog.example.com" {
					t.Errorf("got %s want %s", tr.Requests()[0].Host, want)
				}
			})
		}
	})
}

func TestReplaceDSN(t *testing.T) {
	tests := []struct {
		dsn       string
		hostRules hostRules
		want      string
	}{
		{
			"mysql://user:pass@db.example.com:3306/dbname",
			hostRules{
				{"db.example.com", "127.0.0.1"},
			},
			"mysql://user:pass@127.0.0.1:3306/dbname",
		},
		{
			"mysql://user:pass@db.example.com/dbname",
			hostRules{
				{"db.example.com", "127.0.0.1:1234"},
			},
			"mysql://user:pass@127.0.0.1:1234/dbname",
		},
		{
			"mysql://user:pass@db.example.com:3306/dbname",
			hostRules{
				{"*.example.com", "127.0.0.1"},
			},
			"mysql://user:pass@127.0.0.1:3306/dbname",
		},
		{
			"sqlite3:///path/to/db.sqlite3",
			hostRules{
				{"/path/to", "127.0.0.1"},
			},
			"sqlite3:///path/to/db.sqlite3",
		},
		{
			"sqlite3://path/to/db.sqlite3",
			hostRules{
				{"path", "127.0.0.1"},
			},
			"sqlite3://path/to/db.sqlite3",
		},
		{
			"spanner://test-project/test-instance/test-database",
			hostRules{
				{"test-project", "other-project"},
			},
			"spanner://other-project/test-instance/test-database",
		},
	}
	for _, tt := range tests {
		t.Run(tt.dsn, func(t *testing.T) {
			got := tt.hostRules.replaceDSN(tt.dsn)
			if got != tt.want {
				t.Errorf("got %s want %s", got, tt.want)
			}
		})
	}
}

func TestReplaceAddr(t *testing.T) {
	tests := []struct {
		addr      string
		hostRules hostRules
		want      string
	}{
		{
			"ssh.example.com",
			hostRules{
				{"ssh.example.com", "127.0.0.1"},
			},
			"127.0.0.1",
		},
		{
			"ssh.example.com",
			hostRules{},
			"ssh.example.com",
		},
		{
			"ssh.example.com",
			hostRules{
				{"ssh.example.com", "127.0.0.1:2222"},
			},
			"127.0.0.1:2222",
		},
		{
			"ssh.example.com:22",
			hostRules{
				{"ssh.example.com", "127.0.0.1:2222"},
			},
			"127.0.0.1:2222",
		},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			got := tt.hostRules.replaceAddr(tt.addr)
			if got != tt.want {
				t.Errorf("got %s want %s", got, tt.want)
			}
		})
	}
}
