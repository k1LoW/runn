package capture

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"testing"

	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/testutil"
	"github.com/tenntenn/golden"
)

func TestRunbook(t *testing.T) {
	tests := []struct {
		book string
	}{
		{filepath.Join(testutil.Testdata(), "book", "http.yml")},
		{filepath.Join(testutil.Testdata(), "book", "http_multipart.yml")},
		{filepath.Join(testutil.Testdata(), "book", "grpc.yml")},
		{filepath.Join(testutil.Testdata(), "book", "db.yml")},
		{filepath.Join(testutil.Testdata(), "book", "exec.yml")},
		{filepath.Join(testutil.Testdata(), "book", "include_main.yml")},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(filepath.Base(tt.book), func(t *testing.T) {
			dir := t.TempDir()
			hs := testutil.HTTPServer(t)
			gs := testutil.GRPCServer(t, false, false)
			db, _ := testutil.SQLite(t)
			opts := []runn.Option{
				runn.Book(tt.book),
				runn.HTTPRunner("req", hs.URL, hs.Client(), runn.MultipartBoundary(testutil.MultipartBoundary)),
				runn.GrpcRunner("greq", gs.Conn()),
				runn.DBRunner("db", db),
				runn.Capture(Runbook(dir)),
				runn.Scopes(runn.ScopeAllowReadParent),
			}
			o, err := runn.New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}

			got := golden.Txtar(t, dir)
			f := fmt.Sprintf("%s.runbook", filepath.Base(tt.book))
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testutil.Testdata(), f, got)
				return
			}

			if diff := golden.Diff(t, testutil.Testdata(), f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRunnable(t *testing.T) {
	tests := []struct {
		book string
	}{
		{filepath.Join(testutil.Testdata(), "book", "http.yml")},
		{filepath.Join(testutil.Testdata(), "book", "http_multipart.yml")},
		{filepath.Join(testutil.Testdata(), "book", "grpc.yml")},
		{filepath.Join(testutil.Testdata(), "book", "db.yml")},
		{filepath.Join(testutil.Testdata(), "book", "exec.yml")},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(filepath.Base(tt.book), func(t *testing.T) {
			dir := t.TempDir()
			{
				hs := testutil.HTTPServer(t)
				gs := testutil.GRPCServer(t, false, false)
				db, _ := testutil.SQLite(t)
				opts := []runn.Option{
					runn.Book(tt.book),
					runn.HTTPRunner("req", hs.URL, hs.Client(), runn.MultipartBoundary(testutil.MultipartBoundary)),
					runn.GrpcRunner("greq", gs.Conn()),
					runn.DBRunner("db", db),
					runn.Capture(Runbook(dir)),
					runn.Scopes(runn.ScopeAllowReadParent),
				}
				o, err := runn.New(opts...)
				if err != nil {
					t.Fatal(err)
				}
				if err := o.Run(ctx); err != nil {
					t.Error(err)
				}
			}

			{
				hs := testutil.HTTPServer(t)
				gs := testutil.GRPCServer(t, false, false)
				db, _ := testutil.SQLite(t)
				opts := []runn.Option{
					runn.Book(filepath.Join(dir, capturedFilename(tt.book))),
					runn.HTTPRunner("req", hs.URL, hs.Client(), runn.MultipartBoundary(testutil.MultipartBoundary)),
					runn.GrpcRunner("greq", gs.Conn()),
					runn.DBRunner("db", db),
					runn.Scopes(runn.ScopeAllowReadParent),
				}
				o, err := runn.New(opts...)
				if err != nil {
					t.Fatal(err)
				}
				if err := o.Run(ctx); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
