package runn

import (
	"bytes"
	"context"
	"strings"

	"github.com/cli/safeexec"
	"github.com/k1LoW/exec"
)

const execRunnerKey = "exec"

type execRunner struct {
	operator *operator
}

type execCommand struct {
	command string
	stdin   string
}

func newExecRunner(o *operator) (*execRunner, error) {
	return &execRunner{
		operator: o,
	}, nil
}

func (rnr *execRunner) Run(ctx context.Context, c *execCommand) error {
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	rnr.operator.capturers.captureExecCommand(c.command)

	sh, err := safeexec.LookPath("sh")
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, sh, "-c", c.command)
	if strings.Trim(c.stdin, " \n") != "" {
		cmd.Stdin = strings.NewReader(c.stdin)

		rnr.operator.capturers.captureExecStdin(c.stdin)
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	_ = cmd.Run()

	rnr.operator.capturers.captureExecStdout(stdout.String())
	rnr.operator.capturers.captureExecStderr(stderr.String())

	rnr.operator.record(map[string]interface{}{
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": cmd.ProcessState.ExitCode(),
	})
	return nil
}
