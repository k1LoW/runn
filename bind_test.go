package runn

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/k1LoW/runn/internal/store"
)

func TestBindRunnerRun(t *testing.T) {
	tests := []struct {
		store    *store.Store
		bindCond map[string]any
		want     *store.Store
		wantMap  map[string]any
	}{
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				return s
			}(),
			map[string]any{},
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			map[string]any{
				"newkey": "vars.key",
			},
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				if err := s.SetBindVar("newkey", "value"); err != nil {
					t.Fatal(err)
				}
				s.Record(0, map[string]any{})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			map[string]any{
				"newkey": "'hello'",
			},
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				if err := s.SetBindVar("newkey", "hello"); err != nil {
					t.Fatal(err)
				}
				s.Record(0, map[string]any{})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			map[string]any{
				"newkey": []any{"vars.key", 4, "'hello'"},
			},
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				if err := s.SetBindVar("newkey", []any{"value", 4, "hello"}); err != nil {
					t.Fatal(err)
				}
				s.Record(0, map[string]any{})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			map[string]any{
				"newkey": map[string]any{
					"vars.key": "'hello'",
					"key":      "vars.key",
				},
			},
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				if err := s.SetBindVar("newkey", map[string]any{
					"vars.key": "hello",
					"key":      "value",
				}); err != nil {
					t.Fatal(err)
				}
				s.Record(0, map[string]any{})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			map[string]any{
				"foo['hello']":            "'world'",
				"foo[3]":                  "'three'",
				"foo[vars.key]":           "'four'",
				"foo[vars.key][vars.key]": "'five'",
				"bar[]":                   "'six'",
			},
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				if err := s.SetBindVar("foo", map[any]any{
					"hello": "world",
					3:       "three",
					"value": map[any]any{
						"value": "five",
					},
				}); err != nil {
					t.Fatal(err)
				}
				if err := s.SetBindVar("bar", []any{"six"}); err != nil {
					t.Fatal(err)
				}
				s.Record(0, map[string]any{})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				if err := s.SetBindVar("bar", []any{"six"}); err != nil {
					t.Fatal(err)
				}
				return s
			}(),
			map[string]any{
				"bar[]": "'seven'",
			},
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				if err := s.SetBindVar("bar", []any{"six", "seven"}); err != nil {
					t.Fatal(err)
				}
				s.Record(0, map[string]any{})
				return s
			}(),
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
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			o.store = tt.store
			b := newBindRunner()
			s := newStep(0, "stepKey", o, nil)
			s.bindCond = tt.bindCond
			if err := b.Run(ctx, s, true); err != nil {
				t.Fatal(err)
			}

			{
				got := o.store
				opts := []cmp.Option{
					cmp.AllowUnexported(store.Store{}),
					cmpopts.IgnoreFields(store.Store{}, "mr"),
				}
				if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
					t.Error(diff)
				}
			}

			{
				got := o.store.ToMap()
				delete(got, store.RootKeyEnv)
				if diff := cmp.Diff(got, tt.wantMap, nil); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func TestBindRunnerRunError(t *testing.T) {
	tests := []struct {
		bindCond map[string]any
	}{
		{
			map[string]any{
				store.RootKeyVars: "reverved",
			},
		},
		{
			map[string]any{
				store.RootKeySteps: "reverved",
			},
		},
		{
			map[string]any{
				store.RootKeyParent: "reverved",
			},
		},
		{
			map[string]any{
				store.RootKeyIncluded: "reverved",
			},
		},
		{
			map[string]any{
				store.RootKeyCurrent: "reverved",
			},
		},
		{
			map[string]any{
				store.RootKeyPrevious: "reverved",
			},
		},
		{
			map[string]any{
				store.RootKeyCookie: "reverved",
			},
		},
		{
			map[string]any{
				store.RootKeyEnv: "reverved",
			},
		},
		{
			map[string]any{
				store.RootKeyLoopCountIndex: "reverved",
			},
		},
		{
			map[string]any{
				fmt.Sprintf("%s[]", store.RootKeyVars): "reverved",
			},
		},
		{
			map[string]any{
				fmt.Sprintf("%s[3]", store.RootKeyVars): "reverved",
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
