package runn

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/k1LoW/runn/internal/scope"
	"github.com/k1LoW/runn/internal/store"
	"github.com/k1LoW/runn/testutil"
)

func TestIncludeRunnerRun(t *testing.T) {
	tests := []struct {
		path string
		vars map[string]any
		want int
	}{
		{"testdata/book/db.yml", map[string]any{}, 8},
		{"testdata/book/db.yml", map[string]any{"foo": "bar"}, 8},
		{"testdata/book/db.yml", map[string]any{"json": "json://../vars.json"}, 8},
		{"https://raw.githubusercontent.com/k1LoW/runn/main/testdata/book/db.yml", map[string]any{}, 8},
	}
	ctx := context.Background()
	for _, tt := range tests {
		_, dsn := testutil.SQLite(t)
		o, err := New(Runner("db", dsn), Scopes(scope.AllowReadRemote))
		if err != nil {
			t.Fatal(err)
		}
		r, err := newIncludeRunner()
		if err != nil {
			t.Fatal(err)
		}
		s := newStep(0, "stepKey", o, nil)
		s.includeConfig = &includeConfig{path: tt.path, vars: tt.vars}
		if err := r.Run(ctx, s); err != nil {
			t.Fatal(err)
		}

		t.Run("step length", func(t *testing.T) {
			{
				got := o.store.StepLen()
				if want := 1; got != want {
					t.Errorf("got %v\nwant %v", got, want)
				}
			}
			{
				sm := o.store.ToMap()
				sl, ok := sm["steps"].([]map[string]any)
				if !ok {
					t.Fatal("steps not found")
				}
				steps, ok := sl[0]["steps"].([]map[string]any)
				if !ok {
					t.Errorf("failed to cast: %v", sl[0]["steps"])
				}
				got := len(steps)
				if got != tt.want {
					t.Errorf("got %v\nwant %v", got, tt.want)
				}
			}
		})

		t.Run("var length", func(t *testing.T) {
			sm := o.store.ToMap()
			{
				vars, ok := sm["vars"].(map[string]any)
				if !ok {
					t.Errorf("failed to cast: %v", sm["vars"])
				}
				got := len(vars)
				if want := 0; got != want {
					t.Errorf("got %v\nwant %v", got, want)
				}
			}
			{
				sl, ok := sm["steps"].([]map[string]any)
				if !ok {
					t.Fatal("steps not found")
				}
				vars, ok := sl[0]["vars"].(map[string]any)
				if !ok {
					t.Errorf("failed to cast: %v", sl[0]["vars"])
				}
				got := len(vars)
				if want := len(tt.vars); got != want {
					t.Errorf("got %v\nwant %v", got, want)
				}
			}
		})
	}
}

func TestCustomRunner(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"testdata/book/custom_runners.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		ts := testutil.HTTPServer(t)
		t.Setenv("TEST_HTTP_ENDPOINT", ts.URL)
		o, err := New(Book(tt.path))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMultipleIncludeRunnerRun(t *testing.T) {
	tests := []struct {
		path string
		vars map[string]any
	}{
		{
			"testdata/book/multiple_include_a.yml",
			map[string]any{
				"foo":  123,
				"bar":  "123-abc",
				"baz":  "-23",
				"qux":  4,
				"quxx": "2",
				"corge": map[string]any{
					"grault": "1234",
					"garply": 1234,
				},
				"waldo": true,
				"fred":  "false",
			},
		},
		{
			"testdata/book/multiple_include_main.yml",
			map[string]any{
				"foo":  123,
				"bar":  "abc",
				"baz":  100,
				"qux":  -1,
				"quxx": 2,
				"corge": map[string]any{
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
		t.Run(tt.path, func(t *testing.T) {
			o, err := New(Runner("req", "https://example.com"))
			if err != nil {
				t.Fatal(err)
			}
			r, err := newIncludeRunner()
			if err != nil {
				t.Fatal(err)
			}
			s := newStep(0, "stepKey", o, nil)
			s.includeConfig = &includeConfig{path: tt.path, vars: tt.vars}
			if err := r.Run(ctx, s); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestUseParentStore(t *testing.T) {
	host := testutil.CreateHTTPBinContainer(t)
	tests := []struct {
		name        string
		path        string
		parentStore *store.Store
		wantErr     bool
	}{
		{
			"Use parent store in vars: section",
			"testdata/book/use_parent_store_vars.yml",
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("foo", "bar")
				return s
			}(),
			false,
		},
		{
			"Error if there is no parent store",
			"testdata/book/use_parent_store_vars.yml",
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				return s
			}(),
			true,
		},
		{
			"Use parent store in runners: section",
			"testdata/book/use_parent_store_runners.yml",
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("httprunner", host)
				return s
			}(),
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
			r, err := newIncludeRunner()
			if err != nil {
				t.Fatal(err)
			}
			s := newStep(0, "stepKey", o, nil)
			s.includeConfig = &includeConfig{path: tt.path}
			if err := r.Run(ctx, s); err != nil {
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

func TestIncludeVars(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"testdata/book/parent_vars_parent.yml"},
		{"testdata/book/include_vars_main.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(Book(tt.path))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInlineInclude(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"basic", "testdata/book/inline_include.yml"},
		{"with vars", "testdata/book/inline_include_with_vars.yml"},
		{"nested inline->inline", "testdata/book/inline_include_nested.yml"},
		{"inline->path", "testdata/book/inline_include_with_path_include.yml"},
		{"path->inline", "testdata/book/inline_include_from_path.yml"},
		{"with loop", "testdata/book/inline_include_with_loop.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, err := New(Book(tt.path), Scopes(scope.AllowRunExec))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseIncludeConfigInline(t *testing.T) {
	tests := []struct {
		name    string
		v       any
		wantErr bool
		inline  bool
	}{
		{
			"path based",
			map[string]any{"path": "foo.yml"},
			false,
			false,
		},
		{
			"inline steps",
			map[string]any{"steps": []any{map[string]any{"exec": map[string]any{"command": "echo hello"}}}},
			false,
			true,
		},
		{
			"both path and steps",
			map[string]any{"path": "foo.yml", "steps": []any{}},
			true,
			false,
		},
		{
			"neither path nor steps",
			map[string]any{"vars": map[string]any{"a": "b"}},
			true,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parseIncludeConfig(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if c.inline != tt.inline {
				t.Errorf("got inline %v, want %v", c.inline, tt.inline)
			}
		})
	}
}

func TestIncludedRunErr(t *testing.T) {
	dummyErr := errors.New("dummy")
	tests := []struct {
		target error
		want   bool
	}{
		{errors.New("dummy"), false},
		{&includedRunErr{err: dummyErr}, true},
		{fmt.Errorf("dummy: %w", &includedRunErr{err: dummyErr}), true},
		{fmt.Errorf("dummy: %w", fmt.Errorf("dummy: %w", &includedRunErr{err: dummyErr})), true},
		{fmt.Errorf("dummy: %w", fmt.Errorf("dummy: %w", dummyErr)), false},
	}
	for _, tt := range tests {
		if got := errors.Is(&includedRunErr{err: dummyErr}, tt.target); got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}
