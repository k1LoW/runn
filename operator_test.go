package runbk

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		steps []map[string]interface{}
		vars  map[string]string
		in    interface{}
		want  interface{}
	}{
		{
			[]map[string]interface{}{},
			map[string]string{},
			map[string]string{"key": "val"},
			map[string]interface{}{"key": "val"},
		},
		{
			[]map[string]interface{}{},
			map[string]string{"one": "ichi"},
			map[string]string{"key": "{{ vars.one }}"},
			map[string]interface{}{"key": "ichi"},
		},
		{
			[]map[string]interface{}{},
			map[string]string{"one": "ichi"},
			map[string]string{"{{ vars.one }}": "val"},
			map[string]interface{}{"ichi": "val"},
		},
		{
			[]map[string]interface{}{},
			map[string]string{"one": "1"},
			map[string]string{"key": "{{ vars.one }}"},
			map[string]interface{}{"key": uint64(1)},
		},
	}
	for _, tt := range tests {
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		o.store.steps = tt.steps
		o.store.vars = tt.vars

		got, err := o.expand(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/db.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		func() {
			db, err := os.CreateTemp("", "tmp")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(db.Name())
			o, err := New(Book(tt.book), Runner("db", fmt.Sprintf("sqlite://%s", db.Name())))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		}()
	}
}

func TestRunUsingGitHubAPI(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("env GITHUB_TOKEN is not set")
	}
	ctx := context.Background()
	f, err := New(Book("testdata/book/github.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Run(ctx); err != nil {
		t.Error(err)
	}
}
