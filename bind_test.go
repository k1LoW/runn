package runn

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBindRunnerRun(t *testing.T) {
	tests := []struct {
		store   store
		cond    map[string]any
		want    store
		wantMap map[string]any
	}{
		{
			store{
				steps: []map[string]any{},
				vars:  map[string]any{},
			},
			map[string]any{},
			store{
				steps: []map[string]any{
					{"run": true},
				},
				vars: map[string]any{},
			},
			map[string]any{
				"steps": []map[string]any{
					{"run": true},
				},
				"vars": map[string]any{},
			},
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{},
			},
			map[string]any{
				"newkey": "vars.key",
			},
			store{
				steps: []map[string]any{
					{"run": true},
				},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{
					"newkey": "value",
				},
			},
			map[string]any{
				"steps": []map[string]any{
					{"run": true},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"newkey": "value",
			},
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{},
			},
			map[string]any{
				"newkey": "'hello'",
			},
			store{
				steps: []map[string]any{
					{"run": true},
				},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{
					"newkey": "hello",
				},
			},
			map[string]any{
				"steps": []map[string]any{
					{"run": true},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"newkey": "hello",
			},
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{},
			},
			map[string]any{
				"newkey": []any{"vars.key", 4, "'hello'"},
			},
			store{
				steps: []map[string]any{
					{"run": true},
				},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{
					"newkey": []any{"value", 4, "hello"},
				},
			},
			map[string]any{
				"steps": []map[string]any{
					{"run": true},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"newkey": []any{"value", 4, "hello"},
			},
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{},
			},
			map[string]any{
				"newkey": map[string]any{
					"vars.key": "'hello'",
					"key":      "vars.key",
				},
			},
			store{
				steps: []map[string]any{
					{"run": true},
				},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{
					"newkey": map[string]any{
						"vars.key": "hello",
						"key":      "value",
					},
				},
			},
			map[string]any{
				"steps": []map[string]any{
					{"run": true},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"newkey": map[string]any{
					"vars.key": "hello",
					"key":      "value",
				},
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		o.store = tt.store
		b, err := newBindRunner(o)
		if err != nil {
			t.Fatal(err)
		}
		if err := b.Run(ctx, tt.cond, true); err != nil {
			t.Fatal(err)
		}

		{
			got := b.operator.store
			opts := []cmp.Option{
				cmp.AllowUnexported(store{}),
			}
			if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
				t.Errorf("%s", diff)
			}
		}

		{
			got := b.operator.store.toMap()
			delete(got, storeEnvKey)
			if diff := cmp.Diff(got, tt.wantMap, nil); diff != "" {
				t.Errorf("%s", diff)
			}
		}
	}
}

func TestBindRunnerRunError(t *testing.T) {
	tests := []struct {
		cond map[string]any
	}{
		{
			map[string]any{
				storeVarsKey: "reverved",
			},
		},
		{
			map[string]any{
				storeStepsKey: "reverved",
			},
		},
		{
			map[string]any{
				storeParentKey: "reverved",
			},
		},
		{
			map[string]any{
				storeIncludedKey: "reverved",
			},
		},
		{
			map[string]any{
				storeCurrentKey: "reverved",
			},
		},
		{
			map[string]any{
				storePreviousKey: "reverved",
			},
		},
		{
			map[string]any{
				loopCountVarKey: "reverved",
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		b, err := newBindRunner(o)
		if err != nil {
			t.Fatal(err)
		}
		if err := b.Run(ctx, tt.cond, true); err == nil {
			t.Errorf("want error. cond: %v", tt.cond)
		}
	}
}
