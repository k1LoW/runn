package runn

import (
	"testing"
)

func TestPaths(t *testing.T) {
	tests := []struct {
		pathp string
		want  int
	}{
		{"testdata/book/book.yml", 1},
		{"testdata/book/runn_*", 4},
		{"testdata/book/book.yml:testdata/book/http.yml", 2},
		{"testdata/book/book.yml:testdata/book/runn_*.yml", 5},
		{"testdata/book/book.yml:testdata/book/book.yml", 1},
		{"testdata/book/testdata/book/runn_0_success.yml:testdata/book/runn_*.yml", 4},
	}
	for _, tt := range tests {
		t.Run(tt.pathp, func(t *testing.T) {
			paths, err := Paths(tt.pathp)
			if err != nil {
				t.Error(err)
			}
			got := len(paths)
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}
