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
	CACert               string `yaml:"cacert,omitempty"`
	Cert                 string `yaml:"cert,omitempty"`
	Key                  string `yaml:"key,omitempty"`
	SkipVerify           bool   `yaml:"skipVerify,omitempty"`
	Timeout              string `yaml:"timeout,omitempty"`
	UseCookie            *bool  `yaml:"useCookie,omitempty"`
	Trace                *bool  `yaml:"trace,omitempty"`

	openApi3Doc *openapi3.T
}

type grpcRunnerConfig struct {
	Addr        string   `yaml:"addr"`
	TLS         *bool    `yaml:"tls,omitempty"`
	CACert      string   `yaml:"cacert,omitempty"`
	Cert        string   `yaml:"cert,omitempty"`
	Key         string   `yaml:"key,omitempty"`
	SkipVerify  bool     `yaml:"skipVerify,omitempty"`
	ImportPaths []string `yaml:"importPaths,omitempty"`
	Protos      []string `yaml:"protos,omitempty"`

	cacert []byte
	cert   []byte
	key    []byte
}

type sshRunnerConfig struct {
	SSHConfig           string       `yaml:"sshConfig,omitempty"`
	Host                string       `yaml:"host,omitempty"`
	Hostname            string       `yaml:"hostname,omitempty"`
	User                string       `yaml:"user,omitempty"`
	Port                int          `yaml:"port,omitempty"`
	IdentityFile        string       `yaml:"identityFile,omitempty"`
	IdentityKey         string       `yaml:"identityKey,omitempty"`
	KeepSession         bool         `yaml:"keepSession,omitempty"`
	LocalForward        string       `yaml:"localForward,omitempty"`
	KeyboardInteractive []*sshAnswer `yaml:"keyboardInteractive,omitempty"`
}

type sshAnswer struct {
	Match  string `yaml:"match"`
	Answer string `yaml:"answer"`
}

type httpRunnerOption func(*httpRunnerConfig) error

type grpcRunnerOption func(*grpcRunnerConfig) error

type sshRunnerOption func(*sshRunnerConfig) error

func (c *sshRunnerConfig) validate() error {
	if c.Host == "" && c.Hostname == "" {
		return fmt.Errorf("host or hostname is required")
	}
	if c.IdentityFile != "" && c.IdentityKey != "" {
		return fmt.Errorf("identityFile and identityKey cannot be used at the same time")
	}
	return nil
}

// OpenApi3 sets OpenAPI Document using file path.
func OpenApi3(l string) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.OpenApi3DocLocation = l
		return nil
	}
}

// OpenApi3FromData sets OpenAPI Document from data.
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

// SkipValidateRequest sets whether to skip validation of HTTP request with OpenAPI Document.
func SkipValidateRequest(skip bool) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.SkipValidateRequest = skip
		return nil
	}
}

// SkipValidateResponse sets whether to skip validation of HTTP response with OpenAPI Document.
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

func HTTPCACert(path string) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.CACert = path
		return nil
	}
}

func HTTPCert(path string) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.Cert = path
		return nil
	}
}

func HTTPKey(path string) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.Key = path
		return nil
	}
}

func HTTPSkipVerify(skip bool) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.SkipVerify = skip
		return nil
	}
}

func HTTPTimeout(timeout string) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.Timeout = timeout
		return nil
	}
}

func UseCookie(use bool) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.UseCookie = &use
		return nil
	}
}

func HTTPTrace(trace bool) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.Trace = &trace
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

// Protos append protos.
func Protos(protos []string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.Protos = unique(append(c.Protos, protos...))
		return nil
	}
}

// ImportPaths set import paths.
func ImportPaths(paths []string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.ImportPaths = unique(append(c.ImportPaths, paths...))
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

func IdentityKey(k []byte) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.IdentityKey = string(k)
		return nil
	}
}

func KeepSession(enable bool) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.KeepSession = enable
		return nil
	}
}

func LocalForward(l string) sshRunnerOption {
	return func(c *sshRunnerConfig) error {
		c.LocalForward = l
		return nil
	}
}
