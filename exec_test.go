package runn

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cli/safeexec"
	"github.com/google/go-cmp/cmp"
)

func TestExecRun(t *testing.T) {
	if err := setScopes(ScopeAllowRunExec); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := setScopes(ScopeDenyRunExec); err != nil {
			t.Fatal(err)
		}
	})
	tests := []struct {
		command string
		stdin   string
		shell   string
		want    map[string]any
	}{
		{"echo hello!!", "", "", map[string]any{
			"stdout":    "hello!!\n",
			"stderr":    "",
			"exit_code": 0,
			"run":       true,
		}},
		{"cat", "hello!!", "", map[string]any{
			"stdout":    "hello!!",
			"stderr":    "",
			"exit_code": 0,
			"run":       true,
		}},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			r := newExecRunner()
			s := newStep(0, "stepKey", o)
			c := &execCommand{command: tt.command, stdin: tt.stdin, shell: tt.shell}
			if err := r.run(ctx, c, s); err != nil {
				t.Error(err)
				return
			}
			got := o.store.steps[0]
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestExecShell(t *testing.T) {
	if err := setScopes(ScopeAllowRunExec); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := setScopes(ScopeDenyRunExec); err != nil {
			t.Fatal(err)
		}
	})
	tests := []struct {
		shell string
		want  string
	}{
		{"", execDefaultShell},
		{"bash", "bash"},
		{"sh", "sh"},
	}
	ctx := context.Background()
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			r := newExecRunner()
			s := newStep(0, "stepKey", o)
			c := &execCommand{command: "echo $0", shell: tt.shell}
			if err := r.run(ctx, c, s); err != nil {
				t.Error(err)
				return
			}
			want, err := safeexec.LookPath(tt.want)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := o.store.steps[0]["stdout"].(string)
			if !ok {
				t.Fatal("stdout is not string")
			}
			if !strings.HasPrefix(got, want) {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}
