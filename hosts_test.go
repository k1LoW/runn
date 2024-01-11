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
		book := "testdata/book/http_with_host_rules.yml"
		ts, tr := testutil.HTTPServerAndRouter(t)
		t.Setenv("TEST_HTTP_HOST_RULE", strings.TrimPrefix(ts.URL, "http://"))
		o, err := New(Book(book))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
		want := "example.com"
		if tr.Requests()[0].Host != "example.com" {
			t.Errorf("got %s want %s", tr.Requests()[0].Host, want)
		}
	})

	t.Run("gRPC", func(t *testing.T) {
		book := "testdata/book/grpc_with_host_rules.yml"
		ts := testutil.GRPCServer(t, true, false)
		t.Setenv("TEST_GRPC_HOST_RULE", ts.Addr())
		o, err := New(Book(book))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
	})

	t.Run("CDP", func(t *testing.T) {
		book := "testdata/book/cdp_with_host_rules.yml"
		ts, tr := testutil.HTTPServerAndRouter(t)
		t.Setenv("TEST_HTTP_HOST_RULE", strings.TrimPrefix(ts.URL, "http://"))
		o, err := New(Book(book))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
		want := "example.com"
		if tr.Requests()[0].Host != "example.com" {
			t.Errorf("got %s want %s", tr.Requests()[0].Host, want)
		}
	})
}
