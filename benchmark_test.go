package runn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/k1LoW/runn/testutil"
)

func BenchmarkRun(b *testing.B) { //nostyle:repetition
	ctx := context.Background()
	ts := testutil.HTTPServer(b)
	book := "testdata/book/http.yml"
	for i := 0; i < b.N; i++ {
		o, err := New(Book(book), HTTPRunner("req", ts.URL, ts.Client()))
		if err != nil {
			b.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkProfileEnable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRunbookWithProfile(false)
	}
}

func BenchmarkProfileDisable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runRunbookWithProfile(true)
	}
}

func runRunbookWithProfile(disableProfile bool) {
	ctx := context.Background()

	db, err := os.CreateTemp("", "tmp")
	if err != nil {
		panic(err)
	}
	defer os.Remove(db.Name())

	opts := []Option{
		Book("testdata/book/db.yml"),
		Book("testdata/book/only_if_included.yml"),
		Book("testdata/book/if.yml"),
		Book("testdata/book/include_main.yml"),
		DisableProfile(disableProfile),
		Runner("db", fmt.Sprintf("sqlite://%s", db.Name())),
		Scopes(ScopeAllowRunExec),
	}
	o, err := New(opts...)
	if err != nil {
		panic(err)
	}
	if err := o.Run(ctx); err != nil {
		panic(err)
	}
	if !disableProfile {
		buf := new(bytes.Buffer)
		if err := o.DumpProfile(buf); err != nil {
			panic(err)
		}
	}
}
