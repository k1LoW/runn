package runn

import (
	"context"
	"strings"
	"testing"

	"github.com/k1LoW/runn/internal/scope"
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

// BenchmarkOpenAPI3 is a benchmark with OpenAPI.
func BenchmarkOpenAPI3(b *testing.B) {
	const (
		bookCount = 10
		stepCount = 10
	)
	runBenchmarkWithOpenAPI3(b, bookCount, stepCount)
}

func runBenchmark(b *testing.B, bookCount, stepCount, bodySize int) {
	ctx := context.Background()
	body := "data: " + strings.Repeat("a", bodySize)
	ts, pathp := testutil.BenchmarkSet(b, bookCount, stepCount, body)

	for b.Loop() {
		opts := []Option{
			HTTPRunner("req", ts.URL, ts.Client()),
			Scopes(scope.AllowReadParent),
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

func runBenchmarkWithOpenAPI3(b *testing.B, bookCount, stepCount int) {
	ctx := context.Background()
	_, pathp := testutil.BenchmarkSetWithOpenAPI3(b, bookCount, stepCount)

	for b.Loop() {
		opts := []Option{
			Scopes(scope.AllowReadParent),
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
