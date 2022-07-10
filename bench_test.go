package runn

import (
	"testing"
)

func BenchmarkLoad(b *testing.B) {
	opts := []Option{
		Runner("req", "https://api.github.com"),
		Runner("db", "sqlite://path/to/test.db"),
	}
	for i := 0; i < b.N; i++ {
		for j := 0; j < 5; j++ {
			if _, err := Load("testdata/book/**/*", opts...); err != nil {
				b.Error(err)
			}
			if _, err := Load("testdata/book/**/*", opts...); err != nil {
				b.Error(err)
			}
		}
	}
}
