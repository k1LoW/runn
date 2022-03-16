package runn

import (
	"testing"
	"time"
)

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

func TestIntarval(t *testing.T) {
	tests := []struct {
		d       time.Duration
		wantErr bool
	}{
		{1 * time.Second, false},
		{-1 * time.Second, true},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := Interval(tt.d)
		if err := opt(bk); err != nil {
			if !tt.wantErr {
				t.Errorf("got error %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
	}
}
