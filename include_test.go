package runn

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestIncludeRunnerRun(t *testing.T) {
	tests := []struct {
		book string
		want int
	}{
		{"testdata/book/db.yml", 8},
	}
	ctx := context.Background()
	for _, tt := range tests {
		db, err := os.CreateTemp("", "tmp")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(db.Name())
		o, err := New(Runner("db", fmt.Sprintf("sqlite://%s", db.Name())))
		if err != nil {
			t.Fatal(err)
		}
		r, err := newIncludeRunner(o)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Run(ctx, tt.book); err != nil {
			t.Fatal(err)
		}

		{
			got := len(r.operator.store.steps)
			if want := 1; got != want {
				t.Errorf("got %v\nwant %v", got, want)
			}
		}
		{
			got := len(r.operator.store.steps[0]["steps"].([]map[string]interface{}))
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		}
	}
}
