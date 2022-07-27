package runn

import (
	"os"
	"strconv"
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

func TestLoadBook(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"testdata/book/env.yml"},
	}
	debug := false
	os.Setenv("DEBUG", strconv.FormatBool(debug))
	for _, tt := range tests {
		o, err := LoadBook(tt.path)
		if err != nil {
			t.Fatal(err)
		}
		if want := debug; o.Debug != want {
			t.Errorf("got %v\nwant %v", o.Debug, want)
		}
		if want := "5"; o.Interval != want {
			t.Errorf("got %v\nwant %v", o.Interval, want)
		}
	}
}
