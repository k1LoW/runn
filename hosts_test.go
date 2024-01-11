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
