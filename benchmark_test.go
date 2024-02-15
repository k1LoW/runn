package runn

import (
	"context"
	"strings"
	"testing"

	"github.com/k1LoW/runn/testutil"
)

// BenchmarkSingleRunbook is a benchmark of a single runbook.
func BenchmarkSingleRunbook(b *testing.B) {
	const (
		bookCount = 1
		stepCount = 100
		bodySize  = 1000
	)
	runBenchmark(b, bookCount, stepCount, bodySize)
}

// BenchmarkManyRunbooks is a benchmark of many runbooks.
func BenchmarkManyRunbooks(b *testing.B) {
	const (
		bookCount = 1000
		stepCount = 10
		bodySize  = 100
	)
	runBenchmark(b, bookCount, stepCount, bodySize)
}

func runBenchmark(b *testing.B, bookCount, stepCount, bodySize int) {
	ctx := context.Background()
	body := "data: " + strings.Repeat("a", bodySize)
	ts, pathp := testutil.BenchmarkSet(b, bookCount, stepCount, body)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts := []Option{
			HTTPRunner("req", ts.URL, ts.Client()),
			Scopes(ScopeAllowReadParent),
		}
		o, err := Load(pathp, opts...)
		if err != nil {
			b.Fatal(err)
		}
		if err := o.RunN(ctx); err != nil {
			b.Error(err)
		}
	}
}
