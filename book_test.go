package runn

import (
	"testing"
)

func TestNew(t *testing.T) {
	path := "testdata/book/book.yml"
	o, err := New(Book(path))
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
