package runn

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Songmu/prompter"
	"github.com/k1LoW/sshc/v4"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
)

const sshOutTimeout = 1 * time.Second

const (
	sshStoreStdoutKey = "stdout"
	sshStoreStderrKey = "stderr"
)

type sshRunner struct {
	name         string
	addr         string
	client       *ssh.Client
	sess         *ssh.Session
	stdin        io.WriteCloser
	stdout       chan string
	stderr       chan string
	keepSession  bool
	localForward *sshLocalForward
	sessCancel   context.CancelFunc
}

type sshLocalForward struct {
	local  string
	remote string
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
	var opts []sshc.Option
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
	opts = append(opts, sshc.AuthMethod(sshKeyboardInteractive(nil)))

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
	ctx, cancel := context.WithCancel(context.Background())
	rnr.sessCancel = cancel

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

	// local forward
	if rnr.localForward != nil {
		// remote
		local, err := net.Listen("tcp", rnr.localForward.local)
		if err != nil {
			return err
		}

		go func() {
			for {
				lc, err := local.Accept()
				if err != nil {
					log.Println(err)
					break
				}
				rc, err := rnr.client.Dial("tcp", rnr.localForward.remote)
				if err != nil {
					log.Println(err)
					break
				}
				go func() {
					if err := handleConns(ctx, lc, rc); err != nil {
						log.Println(err)
					}
				}()
			}
		}()
	}

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
	if rnr.sessCancel != nil {
		rnr.sessCancel()
	}
	rnr.sess = nil
	rnr.stdin = nil
	rnr.stdout = nil
	rnr.stderr = nil
	rnr.sessCancel = nil
	return nil
}

func (rnr *sshRunner) Close() error {
	return rnr.closeSession()
}

func (rnr *sshRunner) Run(ctx context.Context, s *step) error {
	o := s.parent
	cmd, err := parseSSHCommand(s.sshCommand, o.expandBeforeRecord)
	if err != nil {
		return fmt.Errorf("invalid ssh command: %w", err)
	}
	if err := rnr.run(ctx, cmd, s); err != nil {
		return err
	}
	return nil
}

func (rnr *sshRunner) run(ctx context.Context, c *sshCommand, s *step) error {
	o := s.parent
	if !rnr.keepSession {
		return rnr.runOnce(ctx, c, s)
	}

	o.capturers.captureSSHCommand(c.command)
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

	o.capturers.captureSSHStdout(stdout)
	o.capturers.captureSSHStderr(stderr)

	o.record(map[string]any{
		string(sshStoreStdoutKey): stdout,
		string(sshStoreStderrKey): stderr,
	})
	return nil
}

func (rnr *sshRunner) runOnce(ctx context.Context, c *sshCommand, s *step) error {
	o := s.parent
	o.capturers.captureSSHCommand(c.command)
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

	_ = rnr.sess.Run(c.command)

	o.capturers.captureSSHStdout(stdout.String())
	o.capturers.captureSSHStderr(stderr.String())

	o.record(map[string]any{
		string(sshStoreStdoutKey): stdout.String(),
		string(sshStoreStderrKey): stderr.String(),
	})

	return nil
}

func handleConns(ctx context.Context, lc, rc net.Conn) (err error) {
	defer func() {
		if errr := rc.Close(); errr != nil {
			err = errr
		}
		if errr := lc.Close(); errr != nil {
			err = errr
		}
	}()

	eg, _ := errgroup.WithContext(ctx) // FIXME: context handling
	done := make(chan struct{})

	// remote -> local
	eg.Go(func() error {
		_, err := io.Copy(lc, rc)
		if err != nil {
			return err
		}
		done <- struct{}{}
		return nil
	})

	// local -> remote
	eg.Go(func() error {
		_, err := io.Copy(rc, lc)
		if err != nil {
			return err
		}
		done <- struct{}{}
		return nil
	})

	<-done
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func sshKeyboardInteractive(as []*sshAnswer) ssh.AuthMethod {
	return ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		var answers []string
	L:
		for _, q := range questions {
			if len(as) == 0 {
				answers = append(answers, prompter.Prompt(q, ""))
			} else {
				for _, a := range as {
					if a.Match == "" {
						return nil, errors.New("match: should not be empty")
					}
					re, err := regexp.Compile(a.Match)
					if err != nil {
						return nil, err
					}
					if re.MatchString(q) {
						answers = append(answers, a.Answer)
						continue L
					}
				}
				answers = append(answers, "")
			}
		}
		return answers, nil
	})
}
