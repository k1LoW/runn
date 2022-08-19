package runn

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBindRunnerRun(t *testing.T) {
	tests := []struct {
		store   store
		cond    map[string]string
		want    store
		wantMap map[string]interface{}
	}{
		{
			store{
				steps: []map[string]interface{}{},
				vars:  map[string]interface{}{},
			},
			map[string]string{},
			store{
				steps: []map[string]interface{}{},
				vars:  map[string]interface{}{},
			},
			map[string]interface{}{
				"steps": []map[string]interface{}{},
				"vars":  map[string]interface{}{},
			},
		},
		{
			store{
				steps: []map[string]interface{}{},
				vars: map[string]interface{}{
					"key": "value",
				},
				bindVars: map[string]interface{}{},
			},
			map[string]string{
				"newkey": "vars.key",
			},
			store{
				steps: []map[string]interface{}{},
				vars: map[string]interface{}{
					"key": "value",
				},
				bindVars: map[string]interface{}{
					"newkey": "value",
				},
			},
			map[string]interface{}{
				"steps": []map[string]interface{}{},
				"vars": map[string]interface{}{
					"key": "value",
				},
				"newkey": "value",
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
		if err := b.Run(ctx, tt.cond); err != nil {
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
			if diff := cmp.Diff(got, tt.wantMap, nil); diff != "" {
				t.Errorf("%s", diff)
			}
		}
	}
}
