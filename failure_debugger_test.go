package runn

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/maskedio"
	"github.com/k1LoW/runn/internal/scope"
)

func TestDebugOnFailure(t *testing.T) {
	tests := []struct {
		name        string
		runbook     string
		wantErr     bool
		contains    []string
		notContains []string
	}{
		{
			name: "success",
			runbook: `desc: success
steps:
  command:
    exec:
      command: |
        echo success-stdout
        echo success-stderr >&2
    test: current.exit_code == 0
`,
			notContains: []string{"success-stdout", "success-stderr", "FAILURE DIAGNOSTICS"},
		},
		{
			name: "only failed step",
			runbook: `desc: failure
steps:
  successful_command:
    exec:
      command: |
        echo successful-step-stdout
        echo successful-step-stderr >&2
    test: current.exit_code == 0
  failed_command:
    exec:
      command: |
        echo failed-step-stdout
        echo failed-step-stderr >&2
        exit 7
    test: current.exit_code == 0
`,
			wantErr: true,
			contains: []string{
				"-----START FAILURE DIAGNOSTICS-----",
				"Step: failed_command",
				"Runner: exec (exec)",
				"failed-step-stdout",
				"failed-step-stderr",
				"-----END FAILURE DIAGNOSTICS-----",
			},
			notContains: []string{"successful-step-stdout", "successful-step-stderr"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeFailureDebugRunbook(t, "runbook.yml", tt.runbook)
			var stderr bytes.Buffer
			o, err := New(
				Book(path),
				DebugOnFailure(true),
				Stderr(&stderr),
				Scopes(scope.AllowRunExec, scope.AllowReadParent),
			)
			if err != nil {
				t.Fatal(err)
			}
			err = o.Run(context.Background())
			if tt.wantErr && err == nil {
				t.Fatal("expected run error")
			}
			if !tt.wantErr && err != nil {
				t.Fatal(err)
			}

			got := stderr.String()
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("output does not contain %q:\n%s", want, got)
				}
			}
			for _, notWant := range tt.notContains {
				if strings.Contains(got, notWant) {
					t.Errorf("output unexpectedly contains %q:\n%s", notWant, got)
				}
			}
		})
	}
}

func TestDebugTakesPrecedenceOverDebugOnFailure(t *testing.T) {
	path := writeFailureDebugRunbook(t, "runbook.yml", `desc: debug precedence
steps:
  failed_command:
    exec:
      command: |
        echo precedence-stderr >&2
        exit 1
    test: current.exit_code == 0
`)
	var stderr bytes.Buffer
	o, err := New(
		Book(path),
		Debug(true),
		DebugOnFailure(true),
		Stderr(&stderr),
		Scopes(scope.AllowRunExec, scope.AllowReadParent),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Run(context.Background()); err == nil {
		t.Fatal("expected run error")
	}

	got := stderr.String()
	if strings.Contains(got, "FAILURE DIAGNOSTICS") {
		t.Errorf("failure diagnostics should be disabled by debug:\n%s", got)
	}
	if c := strings.Count(got, "-----START STDERR-----"); c != 1 {
		t.Errorf("stderr block count = %d, want 1:\n%s", c, got)
	}
}

func TestDebugOnFailureMasksSecrets(t *testing.T) {
	const secret = "failure-debug-secret"
	path := writeFailureDebugRunbook(t, "runbook.yml", `desc: mask secrets
vars:
  secret: failure-debug-secret
secrets:
  - vars.secret
steps:
  failed_command:
    exec:
      command: |
        echo '{{ vars.secret }}' >&2
        exit 1
    test: current.exit_code == 0
`)
	var stderr bytes.Buffer
	o, err := New(
		Book(path),
		DebugOnFailure(true),
		Stderr(&stderr),
		Scopes(scope.AllowRunExec, scope.AllowReadParent),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Run(context.Background()); err == nil {
		t.Fatal("expected run error")
	}

	got := stderr.String()
	if strings.Contains(got, secret) {
		t.Errorf("secret is not masked:\n%s", got)
	}
	if !strings.Contains(got, "*****") {
		t.Errorf("masked value not found:\n%s", got)
	}
}

func TestDebugOnFailureTruncatesDiagnostics(t *testing.T) {
	var stderr bytes.Buffer
	out := maskedio.NewWriter(&stderr)
	d := newFailureDebugger(out, newFailureDebugCoordinator())
	d.SetCurrentTrails(nil)
	d.CaptureExecStdout(strings.Repeat("x", failureDebugMaxBytes+1))
	d.CaptureResultByStep(nil, &RunResult{
		Path: "runbook.yml",
		StepResults: []*StepResult{{
			Key:        "failed_command",
			RunnerType: RunnerTypeExec,
			RunnerKey:  execRunnerKey,
			Err:        errors.New("failure"),
		}},
	})

	got := stderr.String()
	want := "... diagnostics truncated at 1048576 bytes ..."
	if !strings.Contains(got, want) {
		t.Errorf("truncation marker not found:\n%s", got[len(got)-256:])
	}
}

func TestDebugOnFailureKeepsLastLoopIteration(t *testing.T) {
	path := writeFailureDebugRunbook(t, "runbook.yml", `desc: loop failure
steps:
  retry:
    loop:
      count: 2
      interval: 0
    exec:
      command: |
        echo loop-stdout-{{ i }}
        echo loop-stderr-{{ i }} >&2
        exit 1
      shell: bash
    test: current.exit_code == 0
`)
	var stderr bytes.Buffer
	o, err := New(
		Book(path),
		DebugOnFailure(true),
		Stderr(&stderr),
		Scopes(scope.AllowRunExec, scope.AllowReadParent),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Run(context.Background()); err == nil {
		t.Fatal("expected run error")
	}

	got := stderr.String()
	if !strings.Contains(got, "loop-stdout-1") || !strings.Contains(got, "loop-stderr-1") {
		t.Errorf("last loop output not found:\n%s", got)
	}
	if strings.Contains(got, "loop-stdout-0") || strings.Contains(got, "loop-stderr-0") {
		t.Errorf("earlier loop output found:\n%s", got)
	}
}

func TestDebugOnFailureReportsIncludedLeafOnly(t *testing.T) {
	dir := t.TempDir()
	child := filepath.Join(dir, "child.yml")
	parent := filepath.Join(dir, "parent.yml")
	writeFile(t, child, `desc: child
steps:
  leaf_failure:
    exec:
      command: |
        echo included-leaf-stderr >&2
        exit 1
    test: current.exit_code == 0
`)
	writeFile(t, parent, `desc: parent
steps:
  include_child:
    include:
      path: child.yml
`)

	var stderr bytes.Buffer
	o, err := New(
		Book(parent),
		DebugOnFailure(true),
		Stderr(&stderr),
		Scopes(scope.AllowRunExec, scope.AllowReadParent),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Run(context.Background()); err == nil {
		t.Fatal("expected run error")
	}

	got := stderr.String()
	if c := strings.Count(got, "-----START FAILURE DIAGNOSTICS-----"); c != 1 {
		t.Errorf("diagnostic block count = %d, want 1:\n%s", c, got)
	}
	if !strings.Contains(got, "Runbook: "+normalizePath(child)) || !strings.Contains(got, "Step: leaf_failure") {
		t.Errorf("included leaf not identified:\n%s", got)
	}
	if strings.Contains(got, "Step: include_child") {
		t.Errorf("outer include diagnostics found:\n%s", got)
	}
}

func TestDebugOnFailureDoesNotMixConcurrentRunbooks(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.yml"), failureDebugConcurrentRunbook("concurrent-token-a"))
	writeFile(t, filepath.Join(dir, "b.yml"), failureDebugConcurrentRunbook("concurrent-token-b"))

	var stderr bytes.Buffer
	o, err := Load(
		filepath.Join(dir, "*.yml"),
		DebugOnFailure(true),
		RunConcurrent(true, 2),
		Stderr(&stderr),
		Scopes(scope.AllowRunExec, scope.AllowReadParent),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.RunN(context.Background()); err != nil {
		t.Fatal(err)
	}

	got := stderr.String()
	if c := strings.Count(got, "-----START FAILURE DIAGNOSTICS-----"); c != 2 {
		t.Fatalf("diagnostic block count = %d, want 2:\n%s", c, got)
	}
	for _, block := range failureDiagnosticBlocks(got) {
		hasA := strings.Contains(block, "concurrent-token-a")
		hasB := strings.Contains(block, "concurrent-token-b")
		if hasA == hasB {
			t.Errorf("diagnostic block is missing a token or contains mixed tokens:\n%s", block)
		}
	}
}

func writeFailureDebugRunbook(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	writeFile(t, path, content)
	return path
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func failureDebugConcurrentRunbook(token string) string {
	return `desc: concurrent failure
steps:
  failed_command:
    exec:
      command: |
        echo ` + token + `
        echo ` + token + ` >&2
        exit 1
    test: current.exit_code == 0
`
}

func failureDiagnosticBlocks(out string) []string {
	const start = "-----START FAILURE DIAGNOSTICS-----"
	const end = "-----END FAILURE DIAGNOSTICS-----"
	var blocks []string
	for {
		i := strings.Index(out, start)
		if i < 0 {
			return blocks
		}
		out = out[i+len(start):]
		j := strings.Index(out, end)
		if j < 0 {
			return blocks
		}
		blocks = append(blocks, out[:j])
		out = out[j+len(end):]
	}
}
