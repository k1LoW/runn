package runn

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cli/safeexec"
	"github.com/k1LoW/exec"
)

const execRunnerKey = "exec"

const (
	execStoreStdoutKey   = "stdout"
	execStoreStderrKey   = "stderr"
	execStoreExitCodeKey = "exit_code"
)

const execDefaultShell = "sh"

type execRunner struct{}

type execCommand struct {
	command string
	shell   string
	stdin   string
}

func newExecRunner() *execRunner {
	return &execRunner{}
}

func (rnr *execRunner) Run(ctx context.Context, s *step) error {
	o := s.parent
	e, err := o.expandBeforeRecord(s.execCommand)
	if err != nil {
		return err
	}
	cmd, ok := e.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid exec command: %v", e)
	}
	command, err := parseExecCommand(cmd)
	if err != nil {
		return fmt.Errorf("invalid exec command: %w", err)
	}
	if err := rnr.run(ctx, command, s); err != nil {
		return err
	}
	return nil
}

func (rnr *execRunner) run(ctx context.Context, c *execCommand, s *step) error {
	o := s.parent
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if c.shell == "" {
		c.shell = execDefaultShell
	}
	o.capturers.captureExecCommand(c.command, c.shell)

	sh, err := safeexec.LookPath(c.shell)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, sh, "-c", c.command)
	if strings.Trim(c.stdin, " \n") != "" {
		cmd.Stdin = strings.NewReader(c.stdin)

		o.capturers.captureExecStdin(c.stdin)
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	_ = cmd.Run()

	o.capturers.captureExecStdout(stdout.String())
	o.capturers.captureExecStderr(stderr.String())

	o.record(map[string]any{
		string(execStoreStdoutKey):   stdout.String(),
		string(execStoreStderrKey):   stderr.String(),
		string(execStoreExitCodeKey): cmd.ProcessState.ExitCode(),
	})
	return nil
}
