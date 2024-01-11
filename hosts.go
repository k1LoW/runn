package runn

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type hostRule struct {
	host string
	rule string
}

type hostRules []hostRule

func (rules hostRules) chromedpOpt() chromedp.ExecAllocatorOption {
	var values []string
	for _, rule := range rules {
		values = append(values, fmt.Sprintf("MAP %s %s", rule.host, rule.rule))
	}
	return chromedp.Flag("host-rules", strings.Join(values, ","))
}

// dialContextFunc returns DialContext() for http.Transport.DialContext.
func (rules hostRules) dialContextFunc() func(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		for _, rule := range rules {
			if rule.host == host {
				var rhost, rport string
				if strings.Contains(rule.rule, ":") {
					rhost, rport, err = net.SplitHostPort(rule.rule)
					if err != nil {
						return nil, err
					}
				} else {
					rhost = rule.rule
					rport = port
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(rhost, rport))
			}
		}
		return dialer.DialContext(ctx, network, addr)
	}
}

// contextDialerFunc returns Dialer() for grpc.WithContextDialer.
func (rules hostRules) contextDialerFunc() func(ctx context.Context, address string) (net.Conn, error) {
	dialer := &net.Dialer{} // Same as google.golang.org/grpc@v1.58.3/internal/transport.dial()
	return func(ctx context.Context, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		var network string
		for _, rule := range rules {
			if rule.host == host {
				var rhost, rport string
				if strings.Contains(rule.rule, ":") {
					rhost, rport, err = net.SplitHostPort(rule.rule)
					if err != nil {
						return nil, err
					}
				} else {
					rhost = rule.rule
					rport = port
				}
				address = net.JoinHostPort(rhost, rport)
				network, address = parseDialTarget(address)
				return dialer.DialContext(ctx, network, address)
			}
		}
		network, address = parseDialTarget(address)
		return dialer.DialContext(ctx, network, address)
	}
}

// parseDialTarget returns the network and address to pass to dialer.
// Copy from google.golang.org/grpc@v1.58.3/internal/transport/http_util.go.
func parseDialTarget(target string) (string, string) {
	net := "tcp"
	m1 := strings.Index(target, ":")
	m2 := strings.Index(target, ":/")
	// handle unix:addr which will fail with url.Parse
	if m1 >= 0 && m2 < 0 {
		if n := target[0:m1]; n == "unix" {
			return n, target[m1+1:]
		}
	}
	if m2 >= 0 {
		t, err := url.Parse(target)
		if err != nil {
			return net, target
		}
		scheme := t.Scheme
		addr := t.Path
		if scheme == "unix" {
			if addr == "" {
				addr = t.Host
			}
			return scheme, addr
		}
	}
	return net, target
}
