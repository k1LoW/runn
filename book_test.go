package runbk

import (
	"testing"
)

func TestNew(t *testing.T) {
	path := "testdata/book/book.yml"
	f, err := New(Book(path))
	if err != nil {
		t.Fatal(err)
	}
	if want := 1; len(f.httpRunners) != want {
		t.Errorf("got %v\nwant %v", len(f.httpRunners), want)
	}
	if want := 1; len(f.dbRunners) != want {
		t.Errorf("got %v\nwant %v", len(f.dbRunners), want)
	}
	if want := 5; len(f.steps) != want {
		t.Errorf("got %v\nwant %v", len(f.steps), want)
	}
}
