package runn

import (
	"testing"
)

func TestFetchPaths(t *testing.T) {
	tests := []struct {
		pathp   string
		want    int
		wantErr bool
	}{
		{"testdata/book/book.yml", 1, false},
		{"testdata/book/notexist.yml", 0, true},
		{"testdata/book/runn_*", 4, false},
		{"testdata/book/book.yml:testdata/book/http.yml", 2, false},
		{"testdata/book/book.yml:testdata/book/runn_*.yml", 5, false},
		{"testdata/book/book.yml:testdata/book/book.yml", 1, false},
		{"testdata/book/runn_0_success.yml:testdata/book/runn_*.yml", 4, false},
		{"github://k1LoW/runn/testdata/book/book.yml", 1, false},
		{"github://k1LoW/runn/testdata/book/runn_*", 4, false},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/book.yml", 1, false},
		{"gist://b908ae0721300ca45f4e8b81b6be246d", 1, false},
		{"gist://def6fa739fba3fcf211b018f41630adc", 0, true},
		{"gist://def6fa739fba3fcf211b018f41630adc/book.yml", 1, false},
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
				if !tt.wantErr {
					t.Errorf("got %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("want err")
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
