package runn

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/k1LoW/runn/testutil"
)

func TestIncludeRunnerRun(t *testing.T) {
	tests := []struct {
		path string
		vars map[string]interface{}
		want int
	}{
		{"testdata/book/db.yml", map[string]interface{}{}, 8},
		{"testdata/book/db.yml", map[string]interface{}{"foo": "bar"}, 8},
		{"testdata/book/db.yml", map[string]interface{}{"json": "json://../vars.json"}, 8},
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
				"corge": map[string]interface{}{
					"grault": "1234",
					"garply": 1234,
				},
				"waldo": true,
				"fred":  "false",
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
				"corge": map[string]interface{}{
					"grault": "1234",
					"garply": 1234,
				},
				"waldo": true,
				"fred":  "false",
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
			t.Error(err)
		}
	}
}

func TestUseParentStore(t *testing.T) {
	host := testutil.CreateHTTPBinContainer(t)
	tests := []struct {
		name        string
		path        string
		parentStore store
		wantErr     bool
	}{
		{
			"Use parent store in vars: section",
			"testdata/book/use_parent_store_vars.yml",
			store{
				vars: map[string]interface{}{
					"foo": "bar",
				},
			},
			false,
		},
		{
			"Error if there is no parent store",
			"testdata/book/use_parent_store_vars.yml",
			store{},
			true,
		},
		{
			"Use parent store in runners: section",
			"testdata/book/use_parent_store_runners.yml",
			store{
				vars: map[string]interface{}{
					"httprunner": host,
				},
			},
			false,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := New(Runner("req", "https://example.com"))
			if err != nil {
				t.Fatal(err)
			}
			o.store = tt.parentStore
			r, err := newIncludeRunner(o)
			if err != nil {
				t.Fatal(err)
			}
			c := &includeConfig{path: tt.path}
			if err := r.Run(ctx, c); err != nil {
				if !tt.wantErr {
					t.Error(err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
		})
	}
}
