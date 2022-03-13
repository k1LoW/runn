package runn

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExecRun(t *testing.T) {
	tests := []struct {
		command string
		stdin   string
		want    map[string]interface{}
	}{
		{"echo hello!!", "", map[string]interface{}{
			"stdout":    "hello!!\n",
			"stderr":    "",
			"exit_code": 0,
		}},
		{"cat", "hello!!", map[string]interface{}{
			"stdout":    "hello!!",
			"stderr":    "",
			"exit_code": 0,
		}},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		r, err := newExecRunner(o)
		if err != nil {
			t.Fatal(err)
		}
		c := &execCommand{command: tt.command, stdin: tt.stdin}
		if err := r.Run(ctx, c); err != nil {
			t.Error(err)
			return
		}
		got := o.store.steps[0]
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
