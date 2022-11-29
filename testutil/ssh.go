package testutil

import (
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

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
