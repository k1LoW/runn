package runn

import (
	"context"
	"testing"

	"github.com/k1LoW/runn/testutil"
)

func BenchmarkRun(b *testing.B) {
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
