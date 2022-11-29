package runn

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/k1LoW/sshc/v3"
	"golang.org/x/crypto/ssh"
)

const sshOutTimeout = 500 * time.Millisecond

type sshRunner struct {
	name        string
	addr        string
	client      *ssh.Client
	sess        *ssh.Session
	stdin       io.WriteCloser
	stdout      chan string
	stderr      chan string
	keepSession bool
	operator    *operator
}

type sshCommand struct {
	command string
}

func newSSHRunner(name, addr string) (*sshRunner, error) {
	u, err := url.Parse(fmt.Sprintf("//%s", addr))
	if err != nil {
		return nil, err
	}
	host := u.Hostname()
	opts := []sshc.Option{}
	if u.User.Username() != "" {
		opts = append(opts, sshc.User(u.User.Username()))
	}
	if u.Port() != "" {
		p, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, err
		}
		opts = append(opts, sshc.Port(p))
	}
	client, err := sshc.NewClient(host, opts...)
	if err != nil {
		return nil, err
	}

	rnr := &sshRunner{
		name:   name,
		addr:   addr,
		client: client,
	}

	if rnr.keepSession {
		if err := rnr.startSession(); err != nil {
			return nil, err
		}
	}

	return rnr, nil
}

func (rnr *sshRunner) startSession() error {
	if !rnr.keepSession {
		return errors.New("could not use startSession() when keepSession = false")
	}

	sess, err := rnr.client.NewSession()
	if err != nil {
		return err
	}
	stdin, err := sess.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := sess.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := sess.StderrPipe()
	if err != nil {
		return err
	}
	if err := sess.Shell(); err != nil {
		return err
	}

	ol := make(chan string)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			ol <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		close(ol)
	}()

	el := make(chan string)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			el <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		close(el)
	}()

	rnr.sess = sess
	rnr.stdin = stdin
	rnr.stdout = ol
	rnr.stderr = el
	return nil
}

func (rnr *sshRunner) closeSession() error {
	if rnr.sess == nil {
		return nil
	}
	rnr.sess.Close()
	rnr.sess = nil
	rnr.stdin = nil
	rnr.stdout = nil
	rnr.stderr = nil
	return nil
}

func (rnr *sshRunner) Close() error {
	return rnr.closeSession()
}

func (rnr *sshRunner) Run(ctx context.Context, c *sshCommand) error {
	if !rnr.keepSession {
		return rnr.runOnce(ctx, c)
	}

	rnr.operator.capturers.captureSSHCommand(c.command)
	stdout := ""
	stderr := ""

	if _, err := fmt.Fprintf(rnr.stdin, "%s\n", strings.TrimRight(c.command, "\n")); err != nil {
		return err
	}

	timer := time.NewTimer(0)
L:
	for {
		timer.Reset(sshOutTimeout)
		select {
		case line, ok := <-rnr.stdout:
			if !ok {
				break L
			}
			stdout += fmt.Sprintf("%s\n", line)
		case line, ok := <-rnr.stderr:
			if !ok {
				break L
			}
			stderr += fmt.Sprintf("%s\n", line)
		case <-timer.C:
			break L
		case <-ctx.Done():
			break L
		}
	}

	rnr.operator.capturers.captureSSHStdout(stdout)
	rnr.operator.capturers.captureSSHStderr(stderr)

	rnr.operator.record(map[string]interface{}{
		"stdout": stdout,
		"stderr": stderr,
	})
	return nil
}

func (rnr *sshRunner) runOnce(ctx context.Context, c *sshCommand) error {
	rnr.operator.capturers.captureSSHCommand(c.command)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	sess, err := rnr.client.NewSession()
	if err != nil {
		return err
	}
	sess.Stdout = stdout
	sess.Stderr = stderr
	rnr.sess = sess
	defer func() {
		_ = rnr.closeSession()
	}()

	if err := rnr.sess.Run(c.command); err != nil {
		return err
	}

	rnr.operator.capturers.captureSSHStdout(stdout.String())
	rnr.operator.capturers.captureSSHStderr(stderr.String())

	rnr.operator.record(map[string]interface{}{
		"stdout": stdout.String(),
		"stderr": stderr.String(),
	})

	return nil
}
