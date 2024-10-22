package runn

import (
	"context"
	"fmt"
	"testing"

	"github.com/expr-lang/expr/parser"
	"github.com/google/go-cmp/cmp"
)

func TestBindRunnerRun(t *testing.T) {
	tests := []struct {
		store    store
		bindCond map[string]any
		want     store
		wantMap  map[string]any
	}{
		{
			store{
				steps:    []map[string]any{},
				vars:     map[string]any{},
				bindVars: map[string]any{},
			},
			map[string]any{},
			store{
				steps: []map[string]any{
					{},
				},
				vars:     map[string]any{},
				bindVars: map[string]any{},
			},
			map[string]any{
				"steps": []map[string]any{
					{},
				},
				"vars":   map[string]any{},
				"parent": nil,
				"runn":   map[string]any{"i": 0},
				"needs":  nil,
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
					{},
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
					{},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"newkey": "value",
				"parent": nil,
				"runn":   map[string]any{"i": 0},
				"needs":  nil,
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
					{},
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
					{},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"newkey": "hello",
				"parent": nil,
				"runn":   map[string]any{"i": 0},
				"needs":  nil,
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
					{},
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
					{},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"newkey": []any{"value", 4, "hello"},
				"parent": nil,
				"runn":   map[string]any{"i": 0},
				"needs":  nil,
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
					{},
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
					{},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"newkey": map[string]any{
					"vars.key": "hello",
					"key":      "value",
				},
				"parent": nil,
				"runn":   map[string]any{"i": 0},
				"needs":  nil,
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
				"foo['hello']":            "'world'",
				"foo[3]":                  "'three'",
				"foo[vars.key]":           "'four'",
				"foo[vars.key][vars.key]": "'five'",
				"bar[]":                   "'six'",
			},
			store{
				steps: []map[string]any{
					{},
				},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{
					"foo": map[any]any{
						"hello": "world",
						3:       "three",
						"value": map[any]any{
							"value": "five",
						},
					},
					"bar": []any{"six"},
				},
			},
			map[string]any{
				"steps": []map[string]any{
					{},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"foo": map[any]any{
					"hello": "world",
					3:       "three",
					"value": map[any]any{
						"value": "five",
					},
				},
				"bar":    []any{"six"},
				"parent": nil,
				"runn":   map[string]any{"i": 0},
				"needs":  nil,
			},
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{
					"bar": []any{"six"},
				},
			},
			map[string]any{
				"bar[]": "'seven'",
			},
			store{
				steps: []map[string]any{
					{},
				},
				vars: map[string]any{
					"key": "value",
				},
				bindVars: map[string]any{
					"bar": []any{"six", "seven"},
				},
			},
			map[string]any{
				"steps": []map[string]any{
					{},
				},
				"vars": map[string]any{
					"key": "value",
				},
				"bar":    []any{"six", "seven"},
				"parent": nil,
				"runn":   map[string]any{"i": 0},
				"needs":  nil,
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		o.store = &tt.store
		b := newBindRunner()
		s := newStep(0, "stepKey", o, nil)
		s.bindCond = tt.bindCond
		if err := b.Run(ctx, s, true); err != nil {
			t.Fatal(err)
		}

		{
			got := *o.store
			opts := []cmp.Option{
				cmp.AllowUnexported(store{}),
			}
			if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
				t.Error(diff)
			}
		}

		{
			got := o.store.toMap()
			delete(got, storeRootKeyEnv)
			if diff := cmp.Diff(got, tt.wantMap, nil); diff != "" {
				t.Error(diff)
			}
		}
	}
}

func TestBindRunnerRunError(t *testing.T) {
	tests := []struct {
		bindCond map[string]any
	}{
		{
			map[string]any{
				storeRootKeyVars: "reverved",
			},
		},
		{
			map[string]any{
				storeRootKeySteps: "reverved",
			},
		},
		{
			map[string]any{
				storeRootKeyParent: "reverved",
			},
		},
		{
			map[string]any{
				storeRootKeyIncluded: "reverved",
			},
		},
		{
			map[string]any{
				storeRootKeyCurrent: "reverved",
			},
		},
		{
			map[string]any{
				storeRootKeyPrevious: "reverved",
			},
		},
		{
			map[string]any{
				storeRootKeyCookie: "reverved",
			},
		},
		{
			map[string]any{
				storeRootKeyEnv: "reverved",
			},
		},
		{
			map[string]any{
				storeRootKeyLoopCountIndex: "reverved",
			},
		},
		{
			map[string]any{
				fmt.Sprintf("%s[]", storeRootKeyVars): "reverved",
			},
		},
		{
			map[string]any{
				fmt.Sprintf("%s[3]", storeRootKeyVars): "reverved",
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		b := newBindRunner()
		s := newStep(0, "stepKey", o, nil)
		s.bindCond = tt.bindCond
		if err := b.Run(ctx, s, true); err == nil {
			t.Errorf("want error. cond: %v", tt.bindCond)
		}
	}
}

func TestNodeToMap(t *testing.T) {
	v := "hello"
	tests := []struct {
		in    string
		store map[string]any
		want  map[string]any
	}{
		{
			"foo[3]",
			map[string]any{},
			map[string]any{
				"foo": map[any]any{
					3: v,
				},
			},
		},
		{
			"foo['hello']",
			map[string]any{},
			map[string]any{
				"foo": map[any]any{
					"hello": v,
				},
			},
		},
		{
			"foo['hello'][4]",
			map[string]any{},
			map[string]any{
				"foo": map[any]any{
					"hello": map[any]any{
						4: v,
					},
				},
			},
		},
		{
			"foo[5][4][3]",
			map[string]any{},
			map[string]any{
				"foo": map[any]any{
					5: map[any]any{
						4: map[any]any{
							3: v,
						},
					},
				},
			},
		},
		{
			"foo[key]",
			map[string]any{
				"key": "hello",
			},
			map[string]any{
				"foo": map[any]any{
					"hello": v,
				},
			},
		},
		{
			"foo[key][key2]",
			map[string]any{
				"key":  "hello",
				"key2": "hello2",
			},
			map[string]any{
				"foo": map[any]any{
					"hello": map[any]any{
						"hello2": v,
					},
				},
			},
		},
		{
			"foo[vars.key.key2]",
			map[string]any{
				"vars": map[any]any{
					"key": map[any]any{
						"key2": "hello",
					},
				},
			},
			map[string]any{
				"foo": map[any]any{
					"hello": v,
				},
			},
		},
		{
			"foo",
			map[string]any{},
			map[string]any{
				"foo": v,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			tr, err := parser.Parse(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			got, err := nodeToMap(tr.Node, v, tt.store)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMergeVars(t *testing.T) {
	tests := []struct {
		store map[string]any
		vars  map[string]any
		want  map[string]any
	}{
		{
			map[string]any{
				"key": "one",
			},
			map[string]any{
				"key": "two",
			},
			map[string]any{
				"key": "two",
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
				},
			},
			map[string]any{
				"parent": "two",
			},
			map[string]any{
				"parent": "two",
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
				},
			},
			map[string]any{
				"parent": []any{"two"},
			},
			map[string]any{
				"parent": []any{"two"},
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[string]any{
						"grandchild": "two",
					},
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child2": map[string]any{
						"grandchild": "three",
					},
					"child3": "three",
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child":  "one",
					"child2": map[string]any{"grandchild": "three"},
					"child3": "three",
				},
			},
		},
		{
			map[string]any{},
			map[string]any{
				"parent": map[any]any{
					0: "zero",
				},
			},
			map[string]any{
				"parent": map[any]any{
					0: "zero",
				},
			},
		},
		{
			map[string]any{
				"parent": map[any]any{
					0: "zero",
				},
			},
			map[string]any{
				"parent": map[any]any{
					1: "one",
				},
			},
			map[string]any{
				"parent": map[any]any{
					0: "zero",
					1: "one",
				},
			},
		},
		{
			map[string]any{
				"parent": map[any]any{
					"zero": "zero!",
				},
			},
			map[string]any{
				"parent": map[any]any{
					1: "one!",
				},
			},
			map[string]any{
				"parent": map[any]any{
					"zero": "zero!",
					1:      "one!",
				},
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[string]any{
						"grandchild": "two",
					},
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child2": map[any]any{
						"grandchild3": "three",
					},
					"child3": "three",
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[any]any{
						"grandchild":  "two",
						"grandchild3": "three",
					},
					"child3": "three",
				},
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[any]any{
						"grandchild": "two",
					},
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child2": map[string]any{
						"grandchild3": "three",
					},
					"child3": "three",
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[any]any{
						"grandchild":  "two",
						"grandchild3": "three",
					},
					"child3": "three",
				},
			},
		},
		{
			map[string]any{
				"parent": []any{
					"one",
				},
			},
			map[string]any{
				"parent": []any{
					"two",
				},
			},
			map[string]any{
				"parent": []any{
					"one",
					"two",
				},
			},
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			got := mergeVars(tt.store, tt.vars)
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}
