package runn

import (
	"context"
	"errors"
	"testing"
)

func TestTestRun(t *testing.T) {
	tests := []struct {
		cond    string
		first   bool
		wantErr any
	}{
		{"vars.foo.bar == 'baz'", false, nil},
		{"vars.foo.bar == 'xxx'", false, &condFalseError{}},
		{"steps[0].res.status == 403", false, nil},
		{"current.res.status == 403", false, nil},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.cond, func(t *testing.T) {
			o, err := New(Var("foo", map[string]any{
				"bar": "baz",
			}))
			o.store.steps = []map[string]any{
				{
					"res": map[string]any{
						"status": 403,
					},
				},
			}
			if err != nil {
				t.Fatal(err)
			}
			r := newTestRunner()
			s := newStep(0, "stepKey", o, nil)
			s.testCond = tt.cond
			if err := r.Run(ctx, s, tt.first); err != nil {
				if !errors.As(err, &tt.wantErr) {
					t.Errorf("got %v\nwant %v", err, tt.wantErr)
				}
				return
			}
			if tt.wantErr != nil {
				t.Error("want error\n")
			}
		})
	}
}
