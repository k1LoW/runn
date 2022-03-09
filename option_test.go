package runn

import "testing"

func TestVar(t *testing.T) {
	bk := newBook()

	if len(bk.Vars) != 0 {
		t.Fatalf("got %v\nwant %v", len(bk.Vars), 0)
	}

	opt := Var("key", "value")
	if err := opt(bk); err != nil {
		t.Fatal(err)
	}

	got := bk.Vars["key"].(string)
	want := "value"
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}
