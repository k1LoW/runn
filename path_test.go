package runn

import (
	"testing"
)

func TestFetchPaths(t *testing.T) {
	tests := []struct {
		pathp string
		want  int
	}{
		{"testdata/book/book.yml", 1},
		{"testdata/book/notexist.yml", 0},
		{"testdata/book/runn_*", 4},
		{"testdata/book/book.yml:testdata/book/http.yml", 2},
		{"testdata/book/book.yml:testdata/book/runn_*.yml", 5},
		{"testdata/book/book.yml:testdata/book/book.yml", 1},
		{"testdata/book/testdata/book/runn_0_success.yml:testdata/book/runn_*.yml", 4},
		{"github://k1LoW/runn/testdata/book/book.yml", 1},
		{"github://k1LoW/runn/testdata/book/runn_*", 4},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/book.yml", 1},
	}
	t.Cleanup(func() {
		if err := RemoveCacheDir(); err != nil {
			t.Fatal(err)
		}
	})
	for _, tt := range tests {
		t.Run(tt.pathp, func(t *testing.T) {
			paths, err := fetchPaths(tt.pathp)
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

func TestShortenPath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"path/to/book.yml", "p/t/book.yml"},
		{"book.yml", "book.yml"},
		{"/path/to/book.yml", "/p/t/book.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := ShortenPath(tt.in)
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}

		})
	}
}
