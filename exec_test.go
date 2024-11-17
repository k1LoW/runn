package runn

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cli/safeexec"
	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/donegroup"
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
		command    string
		stdin      string
		shell      string
		background bool
		want       map[string]any
	}{
		{"echo hello!!", "", "", false, map[string]any{
			"stdout":    "hello!!\n",
			"stderr":    "",
			"exit_code": 0,
		}},
		{"cat", "hello!!", "", false, map[string]any{
			"stdout":    "hello!!",
			"stderr":    "",
			"exit_code": 0,
		}},
		{"sleep 1000", "", "", true, map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			r := newExecRunner()
			s := newStep(0, "stepKey", o, nil)
			c := &execCommand{command: tt.command, stdin: tt.stdin, shell: tt.shell, background: tt.background}
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
			s := newStep(0, "stepKey", o, nil)
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

func TestExecRunWithSecrets(t *testing.T) {
	if err := setScopes(ScopeAllowRunExec); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := setScopes(ScopeDenyRunExec); err != nil {
			t.Fatal(err)
		}
	})
	tests := []struct {
		vars       map[string]any
		secrets    []string
		command    string
		liveOutput bool
		want       map[string]any
		wantStdout string
		wantStderr string
	}{
		{
			map[string]any{"message": "hello"},
			[]string{"vars.message"},
			"echo hello!!",
			false,
			map[string]any{
				"stdout":    "hello!!\n",
				"stderr":    "",
				"exit_code": 0,
			},
			"",
			"",
		},
		{
			map[string]any{"message": "hello"},
			[]string{"vars.message"},
			"echo hello!!",
			true,
			map[string]any{
				"stdout":    "hello!!\n",
				"stderr":    "",
				"exit_code": 0,
			},
			"*****!!\n",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			opts := []Option{
				Stdout(stdout),
				Stderr(stderr),
				Secret(tt.secrets...),
			}
			for k, v := range tt.vars {
				opts = append(opts, Var(k, v))
			}

			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			r := newExecRunner()
			s := newStep(0, "stepKey", o, nil)
			c := &execCommand{command: tt.command, liveOutput: tt.liveOutput}
			if err := r.run(ctx, c, s); err != nil {
				t.Error(err)
				return
			}
			got := o.store.steps[0]
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}

			gotStdout := stdout.String()
			if gotStdout != tt.wantStdout {
				t.Errorf("got %s, want %s", gotStdout, tt.wantStdout)
			}
			gotStderr := stderr.String()
			if gotStderr != tt.wantStderr {
				t.Errorf("got %s, want %s", gotStderr, tt.wantStderr)
			}
		})
	}
}
