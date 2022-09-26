package runn

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalAsListedSteps(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/book.yml"},
		{"testdata/book/db.yml"},
		{"testdata/book/exec.yml"},
		{"testdata/book/github.yml"},
		{"testdata/book/loop.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			b, err := os.ReadFile(tt.book)
			if err != nil {
				t.Fatal(err)
			}
			bka := newBook()
			if err := unmarshalAsListedSteps(b, bka); err != nil {
				t.Fatal(err)
			}
			bkb := newBook()
			if err := unmarshalAsListedSteps2(b, bkb); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(bka, bkb, cmp.AllowUnexported(book{})); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func TestUnmarshalAsMappedSteps(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/http.yml"},
		{"testdata/book/grpc.yml"},
		{"testdata/book/github_map.yml"},
		{"testdata/book/vars.yml"},
		{"testdata/book/multiple_include_main.yml"},
		{"testdata/book/multiple_include_a.yml"},
		{"testdata/book/multiple_include_b.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			b, err := os.ReadFile(tt.book)
			if err != nil {
				t.Fatal(err)
			}
			bka := newBook()
			if err := unmarshalAsMappedSteps(b, bka); err != nil {
				t.Fatal(err)
			}
			bkb := newBook()
			if err := unmarshalAsMappedSteps2(b, bkb); err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(bka, bkb, cmp.AllowUnexported(book{})); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
