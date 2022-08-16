package runn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/k1LoW/stopw"
)

func TestProfile(t *testing.T) {
	tests := []struct {
		book    string
		profile bool
		wantErr bool
		depth   int
	}{
		{"testdata/book/db.yml", true, false, 3},
		{"testdata/book/only_if_included.yml", true, false, 1},
		{"testdata/book/if.yml", true, false, 2},
		{"testdata/book/include_main.yml", true, false, 3},
		{"testdata/book/db.yml", false, true, 0},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			db, err := os.CreateTemp("", "tmp")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				os.Remove(db.Name())
			})
			opts := []Option{
				T(t),
				Book(tt.book),
				Profile(tt.profile),
				Runner("db", fmt.Sprintf("sqlite://%s", db.Name())),
				Func("upcase", strings.ToUpper),
			}
			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
			buf := new(bytes.Buffer)
			if err := o.DumpProfile(buf); err != nil {
				if !tt.wantErr {
					t.Errorf("got %v", err)
				}
				return
			}
			if buf.Len() == 0 {
				t.Error("invalid profile")
			}
			if tt.wantErr {
				t.Error("want error")
			}
			got := calcDepth(o.sw.Result())
			if got != tt.depth {
				t.Errorf("got %v\nwant %v", got, tt.depth)
			}
		})
	}
}

func calcDepth(s *stopw.Span) int {
	d := 0
	if len(s.Breakdown) > 0 {
		d += 1
		most := 0
		for _, b := range s.Breakdown {
			d := calcDepth(b)
			if most < d {
				most = d
			}
		}
		d += most
	}
	return d
}
