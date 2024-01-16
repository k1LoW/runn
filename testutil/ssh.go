package testutil

import (
	"io"
	"net"
	"strconv"
	"testing"

	sshd "github.com/gliderlabs/ssh"
	"golang.org/x/crypto/ssh"
)

func SSHServer(t testing.TB) string {
	t.Helper()
	var handler sshd.Handler = func(s sshd.Session) {
		authorizedKey := ssh.MarshalAuthorizedKey(s.PublicKey())
		s.Write(authorizedKey)
	}
	host := "127.0.0.1"
	port := NewPort(t)
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	ts := &sshd.Server{Addr: addr, Handler: handler}
	if err := ts.SetOption(sshd.PublicKeyAuth(func(ctx sshd.Context, key sshd.PublicKey) bool {
		return true // allow all keys, or use ssh.KeysEqual() to compare against known keys
	})); err != nil {
		t.Fatal(err)
	}
	ch := make(chan struct{})
	go func() {
		_ = ts.ListenAndServe()
		close(ch)
	}()
	t.Cleanup(func() {
		if err := ts.Close(); err != nil {
			t.Fatal(err)
		}
		<-ch
	})
	return addr
}

func NewNullSSHClient() *ssh.Client {
	return ssh.NewClient(&NullConn{}, nil, nil)
}

type NullConn struct{}

func (*NullConn) User() string          { return "" }
func (*NullConn) SessionID() []byte     { return nil }
func (*NullConn) ClientVersion() []byte { return nil }
func (*NullConn) ServerVersion() []byte { return nil }
func (*NullConn) RemoteAddr() net.Addr  { return nil }
func (*NullConn) LocalAddr() net.Addr   { return nil }
func (*NullConn) SendRequest(name string, wantReply bool, payload []byte) (bool, []byte, error) {
	return true, nil, nil
}
func (*NullConn) OpenChannel(name string, data []byte) (ssh.Channel, <-chan *ssh.Request, error) {
	return &NullChannel{}, nil, nil
}
func (*NullConn) Close() error { return nil }
func (*NullConn) Wait() error  { return nil }

type NullChannel struct{}

func (*NullChannel) Read(data []byte) (int, error)  { return 10, nil }
func (*NullChannel) Write(data []byte) (int, error) { return 10, nil }
func (*NullChannel) Close() error                   { return nil }
func (*NullChannel) CloseWrite() error              { return nil }
func (*NullChannel) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return true, nil
}
func (*NullChannel) Stderr() io.ReadWriter { return &NullReadWriter{} }

type NullReadWriter struct{}

func (*NullReadWriter) Read(data []byte) (int, error)  { return 10, nil }
func (*NullReadWriter) Write(data []byte) (int, error) { return 10, nil }
