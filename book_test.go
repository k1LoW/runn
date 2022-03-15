package runn

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func TestNew(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"testdata/book/book.yml"},
		{"testdata/book/map.yml"},
	}
	for _, tt := range tests {
		o, err := New(Book(tt.path))
		if err != nil {
			t.Fatal(err)
		}
		if want := 1; len(o.httpRunners) != want {
			t.Errorf("got %v\nwant %v", len(o.httpRunners), want)
		}
		if want := 1; len(o.dbRunners) != want {
			t.Errorf("got %v\nwant %v", len(o.dbRunners), want)
		}
		if want := 6; len(o.steps) != want {
			t.Errorf("got %v\nwant %v", len(o.steps), want)
		}
	}
}
