package runn

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

type httpRunnerConfig struct {
	Endpoint             string `yaml:"endpoint"`
	OpenApi3DocLocation  string `yaml:"openapi3,omitempty"`
	SkipValidateRequest  bool   `yaml:"skipValidateRequest,omitempty"`
	SkipValidateResponse bool   `yaml:"skipValidateResponse,omitempty"`
	NotFollowRedirect    bool   `yaml:"notFollowRedirect,omitempty"`
	MultipartBoundary    string `yaml:"multipartBoundary,omitempty"`

	openApi3Doc *openapi3.T
}

type grpcRunnerConfig struct {
	Addr       string `yaml:"addr"`
	TLS        *bool  `yaml:"tls,omitempty"`
	CACert     string `yaml:"cacert,omitempty"`
	Cert       string `yaml:"cert,omitempty"`
	Key        string `yaml:"key,omitempty"`
	SkipVerify bool   `yaml:"skipVerify,omitempty"`

	cacert []byte
	cert   []byte
	key    []byte
}

type sshRunnerConfig struct {
	SSHConfig    string `yaml:"sshConfig,omitempty"`
	Host         string `yaml:"host,omitempty"`
	Hostname     string `yaml:"hostname,omitempty"`
	User         string `yaml:"user,omitempty"`
	Port         int    `yaml:"port,omitempty"`
	IdentityFile string `yaml:"identityFile,omitempty"`
	KeepSession  bool   `yaml:"keepSession,omitempty"`
}

type httpRunnerOption func(*httpRunnerConfig) error

type grpcRunnerOption func(*grpcRunnerConfig) error

type sshRunnerOption func(*sshRunnerConfig) error

func OpenApi3(l string) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.OpenApi3DocLocation = l
		return nil
	}
}

func OpenApi3FromData(d []byte) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		ctx := context.Background()
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData(d)
		if err != nil {
			return err
		}
		if err := doc.Validate(ctx); err != nil {
			return fmt.Errorf("openapi document validation error: %w", err)
		}
		c.openApi3Doc = doc
		return nil
	}
}

func SkipValidateRequest(skip bool) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.SkipValidateRequest = skip
		return nil
	}
}

func SkipValidateResponse(skip bool) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.SkipValidateResponse = skip
		return nil
	}
}

func NotFollowRedirect(nf bool) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.NotFollowRedirect = nf
		return nil
	}
}

func MultipartBoundary(b string) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.MultipartBoundary = b
		return nil
	}
}

func TLS(useTLS bool) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.TLS = &useTLS
		return nil
	}
}

func CACert(path string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.CACert = path
		return nil
	}
}

func Cert(path string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.Cert = path
		return nil
	}
}

func Key(path string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.Key = path
		return nil
	}
}

func CACertFromData(b []byte) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.cacert = b
		return nil
	}
}

func CertFromData(b []byte) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.cert = b
		return nil
	}
}

func KeyFromData(b []byte) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.key = b
		return nil
	}
}

func SSHConfig(p string) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.SSHConfig = p
		return nil
	}
}

func Host(h string) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.Host = h
		return nil
	}
}

func Hostname(h string) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.Hostname = h
		return nil
	}
}

func User(u string) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.User = u
		return nil
	}
}

func Port(p int) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.Port = p
		return nil
	}
}

func IdentityFile(p string) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.IdentityFile = p
		return nil
	}
}

func KeepSession(enable bool) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.KeepSession = enable
		return nil
	}
}
