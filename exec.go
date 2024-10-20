package runn

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cli/safeexec"
	"github.com/k1LoW/donegroup"
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
	command    string
	shell      string
	stdin      string
	background bool
}

type execResult struct {
	stdout string
	stderr string
	err    error
}

func newExecRunner() *execRunner {
	return &execRunner{}
}

func (rnr *execRunner) Run(ctx context.Context, s *step) error {
	globalScopes.mu.RLock()
	if !globalScopes.runExec {
		globalScopes.mu.RUnlock()
		return errors.New("scope error: exec runner is not allowed. 'run:exec' scope is required")
	}
	globalScopes.mu.RUnlock()
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
	if c.shell == "" {
		c.shell = execDefaultShell
	}
	o.capturers.captureExecCommand(c.command, c.shell, c.background)

	sh, err := safeexec.LookPath(c.shell)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, sh, "-c", c.command)
	if strings.Trim(c.stdin, " \n") != "" {
		cmd.Stdin = strings.NewReader(c.stdin)

		o.capturers.captureExecStdin(c.stdin)
	}

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return fmt.Errorf("error creating StdoutPipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return fmt.Errorf("error creating StderrPipe: %w", err)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting command: %w", err)
	}

	done := make(chan execResult)
	go func() {
		var sob, seb bytes.Buffer
		scanner := bufio.NewScanner(io.TeeReader(stdout, io.MultiWriter(&sob, io.Discard)))
		o.capturers.captureExecStdoutStart(c.command)
		for scanner.Scan() {
			o.capturers.captureExecStdoutLine(scanner.Text())
		}

		o.capturers.captureExecStdoutEnd(c.command)

		if err := scanner.Err(); err != nil {
			done <- execResult{
				"",
				"",
				fmt.Errorf("error reading command output: %w", err),
			}
			return
		}

		sops := sob.String()

		o.capturers.captureExecStdout(sops)

		scanner = bufio.NewScanner(io.TeeReader(stderr, io.MultiWriter(&seb, io.Discard)))
		o.capturers.captureExecStderrStart(c.command)
		for scanner.Scan() {
			o.capturers.captureExecStderrLine(scanner.Text())
		}
		o.capturers.captureExecStderrEnd(c.command)
		if err := scanner.Err(); err != nil {
			done <- execResult{
				sops,
				"",
				fmt.Errorf("error reading command error: %w", err),
			}

			return
		}
		seps := seb.String()

		o.capturers.captureExecStderr(seps)

		done <- execResult{
			sops,
			seps,
			nil,
		}
	}()

	if c.background {
		donegroup.Go(ctx, func() error {
			var result execResult
			select {
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					return fmt.Errorf("command timed out")
				}
				return nil
			case result = <-done:
				if result.err != nil {
					return result.err
				}
			}

			err = cmd.Wait() // WHY: Because it is only necessary to wait. For example, SIGNAL KILL is also normal.
			if err != nil {
				return fmt.Errorf("command finished with error: %w", err)
			}

			o.record(map[string]any{
				string(execStoreStdoutKey):   result.stdout,
				string(execStoreStderrKey):   result.stderr,
				string(execStoreExitCodeKey): cmd.ProcessState.ExitCode(),
			})

			return nil
		})

		o.record(map[string]any{})
		return nil
	}

	var result execResult
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out")
		}
		return fmt.Errorf("command was canceled")
	case result = <-done:
		if result.err != nil {
			return result.err
		}
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("command finished with error: %w", err)
	}

	o.record(map[string]any{
		string(execStoreStdoutKey):   result.stdout,
		string(execStoreStderrKey):   result.stderr,
		string(execStoreExitCodeKey): cmd.ProcessState.ExitCode(),
	})

	return nil
}
