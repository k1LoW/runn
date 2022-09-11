package runn

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func TestIncludeRunnerRun(t *testing.T) {
	tests := []struct {
		path string
		vars map[string]interface{}
		want int
	}{
		{"testdata/book/db.yml", map[string]interface{}{}, 8},
		{"testdata/book/db.yml", map[string]interface{}{"foo": "bar"}, 8},
		{"testdata/book/db.yml", map[string]interface{}{"json": "json://testdata/vars.json"}, 8},
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
		c := &includeConfig{path: tt.path, vars: tt.vars}
		if err := r.Run(ctx, c); err != nil {
			t.Fatal(err)
		}

		t.Run("step length", func(t *testing.T) {
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
		})

		t.Run("var length", func(t *testing.T) {
			{
				got := len(r.operator.store.vars)
				if want := 0; got != want {
					t.Errorf("got %v\nwant %v", got, want)
				}
			}
			{
				got := len(r.operator.store.steps[0]["vars"].(map[string]interface{}))
				if want := len(tt.vars); got != want {
					t.Errorf("got %v\nwant %v", got, want)
				}
			}
		})
	}
}

func TestMultipleIncludeRunnerRun(t *testing.T) {
	tests := []struct {
		path string
		vars map[string]interface{}
	}{
		{
			"testdata/book/multiple_include_a.yml",
			map[string]interface{}{
				"foo":  123,
				"bar":  "123-abc",
				"baz":  "-23",
				"qux":  4,
				"quxx": "2",
			},
		},
		{
			"testdata/book/multiple_include_main.yml",
			map[string]interface{}{
				"foo":  123,
				"bar":  "abc",
				"baz":  100,
				"qux":  -1,
				"quxx": 2,
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(Runner("req", "https://example.com"))
		if err != nil {
			t.Fatal(err)
		}
		r, err := newIncludeRunner(o)
		if err != nil {
			t.Fatal(err)
		}
		c := &includeConfig{path: tt.path, vars: tt.vars}
		if err := r.Run(ctx, c); err != nil {
			t.Fatal(err)
		}
	}
}
