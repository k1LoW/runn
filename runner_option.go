package runn

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/runn/internal/deprecation"
	"github.com/k1LoW/runn/internal/sliceutil"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
)

type httpRunnerConfig struct {
	Endpoint                   string `yaml:"endpoint"`
	OpenAPI3DocLocation        string `yaml:"openapi3,omitempty"`
	SkipValidateRequest        bool   `yaml:"skipValidateRequest,omitempty"`
	SkipValidateResponse       bool   `yaml:"skipValidateResponse,omitempty"`
	SkipCircularReferenceCheck bool   `yaml:"skipCircularReferenceCheck,omitempty"`
	NotFollowRedirect          bool   `yaml:"notFollowRedirect,omitempty"`
	MultipartBoundary          string `yaml:"multipartBoundary,omitempty"`
	CACert                     string `yaml:"cacert,omitempty"`
	Cert                       string `yaml:"cert,omitempty"`
	Key                        string `yaml:"key,omitempty"`
	SkipVerify                 bool   `yaml:"skipVerify,omitempty"`
	Timeout                    string `yaml:"timeout,omitempty"`
	UseCookie                  *bool  `yaml:"useCookie,omitempty"`
	Trace                      traceConfig

	openAPI3Doc libopenapi.Document
}

type traceConfig struct {
	Enable     *bool  `yaml:"enable"`
	HeaderName string `yaml:"headerName,omitempty"`
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
	BufDirs     []string `yaml:"bufDirs,omitempty"`
	BufLocks    []string `yaml:"bufLocks,omitempty"`
	BufConfigs  []string `yaml:"bufConfigs,omitempty"`
	BufModules  []string `yaml:"bufModules,omitempty"`
	Trace       traceConfig

	cacert []byte
	cert   []byte
	key    []byte
}

type dbRunnerConfig struct {
	DSN   string `yaml:"dsn"`
	Trace *bool  `yaml:"trace,omitempty"`
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

type includeRunnerConfig struct {
	Path   string         `yaml:"path"`
	Params map[string]any `yaml:"params,omitempty"`
}

type cdpRunnerConfig struct {
	Addr    string         `yaml:"addr,omitempty"`
	Flags   map[string]any `yaml:"flags,omitempty"`
	Timeout string         `yaml:"timeout,omitempty"`
	Remote  string         `yaml:"-"`
}

type httpRunnerOption func(*httpRunnerConfig) error

type grpcRunnerOption func(*grpcRunnerConfig) error

type dbRunnerOption func(*dbRunnerConfig) error

type sshRunnerOption func(*sshRunnerConfig) error

type cdpRunnerOption func(*cdpRunnerConfig) error

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
// Deprecated: Use OpenAPI3 instead.
func OpenApi3(l string) httpRunnerOption {
	deprecation.AddWarning("OpenApi3", "runn.OpenApi3 is deprecated. Use runn.OpenAPI3 instead.")
	return OpenAPI3(l)
}

// OpenAPI3 sets OpenAPI Document using file path.
func OpenAPI3(l string) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.OpenAPI3DocLocation = l
		return nil
	}
}

// OpenApi3FromData sets OpenAPI Document from data.
// Deprecated: Use OpenAPI3FromData instead.
func OpenApi3FromData(d []byte) httpRunnerOption {
	deprecation.AddWarning("OpenApi3FromData", "runn.OpenApi3FromData is deprecated. Use runn.OpenAPI3FromData instead.")
	return OpenAPI3FromData(d)
}

// OpenAPI3FromData sets OpenAPI Document from data.
func OpenAPI3FromData(d []byte) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		hash := hashBytes(d)
		od, ok := globalOpenAPI3DocRegistory[hash]
		if ok {
			c.openAPI3Doc = od
			return nil
		}
		oc := &datamodel.DocumentConfiguration{
			AllowFileReferences:        true,
			AllowRemoteReferences:      true,
			SkipCircularReferenceCheck: c.SkipCircularReferenceCheck,
		}
		doc, err := libopenapi.NewDocumentWithConfiguration(d, oc)
		if err != nil {
			return err
		}
		c.openAPI3Doc = doc
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

// SkipCircularReferenceCheck sets whether to skip circular reference check in OpenAPI Document.
func SkipCircularReferenceCheck(skip bool) httpRunnerOption {
	return func(c *httpRunnerConfig) error {
		c.SkipCircularReferenceCheck = skip
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
		c.Trace.Enable = &trace
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
		c.Protos = sliceutil.Unique(append(c.Protos, protos...))
		return nil
	}
}

// ImportPaths set import paths.
func ImportPaths(paths []string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.ImportPaths = sliceutil.Unique(append(c.ImportPaths, paths...))
		return nil
	}
}

func GRPCTrace(trace bool) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.Trace.Enable = &trace
		return nil
	}
}

func BufDir(dirs ...string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.BufDirs = sliceutil.Unique(append(c.BufDirs, dirs...))
		return nil
	}
}

func BufLock(locks ...string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.BufLocks = sliceutil.Unique(append(c.BufLocks, locks...))
		return nil
	}
}

func BufConfig(configs ...string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.BufConfigs = sliceutil.Unique(append(c.BufConfigs, configs...))
		return nil
	}
}

func BufModule(modules ...string) grpcRunnerOption {
	return func(c *grpcRunnerConfig) error {
		c.BufModules = sliceutil.Unique(append(c.BufModules, modules...))
		return nil
	}
}

func DBTrace(trace bool) dbRunnerOption {
	return func(c *dbRunnerConfig) error {
		c.Trace = &trace
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

// CDPFlag set chromedp flag.
func CDPFlag(flag string, tf any) cdpRunnerOption {
	return func(c *cdpRunnerConfig) error {
		c.Flags[flag] = tf
		return nil
	}
}

// CDPTimeout sets the timeout for each CDP step.
// The timeout should be a valid duration string (e.g., "30s", "2m", "1m30s").
// Default timeout is 60 seconds if not specified.
func CDPTimeout(timeout string) cdpRunnerOption {
	return func(c *cdpRunnerConfig) error {
		c.Timeout = timeout
		return nil
	}
}

func (t *traceConfig) UnmarshalYAML(b []byte) error {
	if enable, err := strconv.ParseBool(strings.TrimSpace(string(b))); err == nil {
		t.Enable = &enable
		return nil
	}

	type Alias traceConfig
	s := &Alias{}
	if err := yaml.Unmarshal(b, s); err != nil {
		return err
	}

	t.Enable = s.Enable
	t.HeaderName = s.HeaderName

	return nil
}
