package runn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cli/safeexec"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/exec"
	"github.com/k1LoW/runn/internal/scope"
	"github.com/mattn/go-shellwords"
)

const execRunnerKey = "exec"

const (
	execStoreStdoutKey   = "stdout"
	execStoreStderrKey   = "stderr"
	execStoreExitCodeKey = "exit_code"
)

const execDefaultShell = "bash -e -c {0}"
const execSh = "sh -e -c {0}"
const execBash = "bash --noprofile --norc -eo pipefail -c {0}"

type execRunner struct{}

type execCommand struct {
	command    string
	shell      string
	stdin      string
	background bool
	liveOutput bool
}

func newExecRunner() *execRunner {
	return &execRunner{}
}

func (rnr *execRunner) Run(ctx context.Context, s *step) error {
	if !scope.IsRunExecAllowed() {
		return errors.New("scope error: exec runner is not allowed. 'run:exec' scope is required")
	}
	o := s.parent
	e, err := o.expandBeforeRecord(s.execCommand, s)
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
	switch c.shell {
	case "":
		c.shell = execDefaultShell
	case "sh":
		c.shell = execSh
	case "bash":
		c.shell = execBash
	}
	o.capturers.captureExecCommand(c.command, c.shell, c.background)
	if !strings.Contains(c.shell, "{0}") {
		return fmt.Errorf("invalid shell setting. custom shell option requires `{0}`.: %q", c.shell)
	}
	shWithOpts, err := shellwords.Parse(c.shell)
	if err != nil {
		return nil
	}
	for i := range shWithOpts {
		shWithOpts[i] = strings.Replace(shWithOpts[i], "{0}", c.command, 1)
	}

	sh, err := safeexec.LookPath(shWithOpts[0])
	if err != nil {
		if c.shell != "" {
			return err
		}
		// fallback to sh
		fallback, errr := safeexec.LookPath("sh")
		if errr != nil {
			return err
		}
		sh = fallback
	}

	cmd := exec.CommandContext(ctx, sh, shWithOpts[1:]...)
	if strings.Trim(c.stdin, " \n") != "" {
		cmd.Stdin = strings.NewReader(c.stdin)

		o.capturers.captureExecStdin(c.stdin)
	}
	if c.liveOutput {
		cmd.Stdout = io.MultiWriter(stdout, o.maskRule.NewWriter(o.stdout))
		cmd.Stderr = io.MultiWriter(stderr, o.maskRule.NewWriter(o.stderr))
	} else {
		cmd.Stdout = stdout
		cmd.Stderr = stderr
	}

	if c.background {
		// run in background
		if err := cmd.Start(); err != nil {
			o.capturers.captureExecStdout(stdout.String())
			o.capturers.captureExecStderr(stderr.String())
			o.record(s.idx, map[string]any{
				string(execStoreStdoutKey):   stdout.String(),
				string(execStoreStderrKey):   stderr.String(),
				string(execStoreExitCodeKey): cmd.ProcessState.ExitCode(),
			})
			return nil
		}
		donegroup.Go(ctx, func() error {
			_ = cmd.Wait() // WHY: Because it is only necessary to wait. For example, SIGNAL KILL is also normal.
			return nil
		})
		o.record(s.idx, map[string]any{})
		return nil
	}

	_ = cmd.Run()

	o.capturers.captureExecStdout(stdout.String())
	o.capturers.captureExecStderr(stderr.String())

	o.record(s.idx, map[string]any{
		string(execStoreStdoutKey):   stdout.String(),
		string(execStoreStderrKey):   stderr.String(),
		string(execStoreExitCodeKey): cmd.ProcessState.ExitCode(),
	})
	return nil
}
