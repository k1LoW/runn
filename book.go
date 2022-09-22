package runn

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/duration"
	"github.com/k1LoW/expand"
	"github.com/rs/xid"
)

const noDesc = "[No Description]"

type book struct {
	Desc          string                   `yaml:"desc,omitempty"`
	Runners       map[string]interface{}   `yaml:"runners,omitempty"`
	Vars          map[string]interface{}   `yaml:"vars,omitempty"`
	Steps         []map[string]interface{} `yaml:"steps,omitempty"`
	Debug         bool                     `yaml:"debug,omitempty"`
	Interval      string                   `yaml:"interval,omitempty"`
	If            string                   `yaml:"if,omitempty"`
	SkipTest      bool                     `yaml:"skipTest,omitempty"`
	funcs         map[string]interface{}   `yaml:"-"`
	stepKeys      []string
	path          string // runbook file path
	httpRunners   map[string]*httpRunner
	dbRunners     map[string]*dbRunner
	grpcRunners   map[string]*grpcRunner
	profile       bool
	interval      time.Duration
	useMap        bool
	t             *testing.T
	included      bool
	failFast      bool
	skipIncluded  bool
	runMatch      *regexp.Regexp
	runSample     int
	runShardIndex int
	runShardN     int
	runnerErrs    map[string]error
	beforeFuncs   []func() error
	afterFuncs    []func() error
	capturers     capturers
}

func newBook() *book {
	return &book{
		Runners:     map[string]interface{}{},
		Vars:        map[string]interface{}{},
		Steps:       []map[string]interface{}{},
		funcs:       map[string]interface{}{},
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
		grpcRunners: map[string]*grpcRunner{},
		interval:    0 * time.Second,
		runnerErrs:  map[string]error{},
	}
}

type usingMappedSteps struct {
	Desc     string                 `yaml:"desc,omitempty"`
	Runners  map[string]interface{} `yaml:"runners,omitempty"`
	Vars     map[string]interface{} `yaml:"vars,omitempty"`
	Steps    yaml.MapSlice          `yaml:"steps,omitempty"`
	Debug    bool                   `yaml:"debug,omitempty"`
	Interval string                 `yaml:"interval,omitempty"`
	If       string                 `yaml:"if,omitempty"`
	SkipTest bool                   `yaml:"skipTest,omitempty"`
}

func newMapped() usingMappedSteps {
	return usingMappedSteps{
		Runners: map[string]interface{}{},
		Vars:    map[string]interface{}{},
		Steps:   yaml.MapSlice{},
	}
}

func loadBook(in io.Reader) (*book, error) {
	bk := newBook()
	b, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}
	b = expand.ExpandenvYAMLBytes(b)
	if err := unmarshalAsListedSteps(b, bk); err != nil {
		if err := unmarshalAsMappedSteps(b, bk); err != nil {
			return nil, err
		}
	}

	for k, v := range bk.Runners {
		if err := validateRunnerKey(k); err != nil {
			return nil, err
		}
		if err := bk.parseRunner(k, v); err != nil {
			bk.runnerErrs[k] = err
		}
	}

	for i, s := range bk.Steps {
		if err := validateStepKeys(s); err != nil {
			return nil, fmt.Errorf("invalid steps[%d]. %w: %s", i, err, s)
		}
	}

	return bk, nil
}

func unmarshalAsListedSteps(b []byte, bk *book) error {
	if err := yaml.Unmarshal(b, bk); err != nil {
		return err
	}
	if bk.Runners == nil {
		bk.Runners = map[string]interface{}{}
	}
	if bk.Vars == nil {
		bk.Vars = map[string]interface{}{}
	} else {
		// To match behavior with json.Marshal
		b, err := json.Marshal(bk.Vars)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &bk.Vars); err != nil {
			return err
		}
	}
	if bk.Desc == "" {
		bk.Desc = noDesc
	}
	return nil
}

func unmarshalAsMappedSteps(b []byte, bk *book) error {
	m := newMapped()
	if err := yaml.Unmarshal(b, &m); err != nil {
		return err
	}
	bk.useMap = true
	bk.Desc = m.Desc
	bk.Runners = m.Runners
	bk.Vars = m.Vars
	bk.Debug = m.Debug
	if bk.Desc == "" {
		bk.Desc = noDesc
	}
	bk.Interval = m.Interval
	bk.If = m.If
	bk.SkipTest = m.SkipTest

	if bk.Interval != "" {
		d, err := duration.Parse(bk.Interval)
		if err != nil {
			return err
		}
		bk.interval = d
	}

	keys := map[string]struct{}{}
	for _, s := range m.Steps {
		bk.Steps = append(bk.Steps, s.Value.(map[string]interface{}))
		var k string
		switch v := s.Key.(type) {
		case string:
			k = v
		case uint64:
			k = fmt.Sprintf("%d", v)
		default:
			k = fmt.Sprintf("%v", v)
		}
		bk.stepKeys = append(bk.stepKeys, k)
		if _, ok := keys[k]; ok {
			return fmt.Errorf("duplicate step keys: %s", k)
		}
		keys[k] = struct{}{}
	}
	return nil
}

func (bk *book) parseRunner(k string, v interface{}) error {
	delete(bk.runnerErrs, k)
	root, err := bk.generateOperatorRoot()
	if err != nil {
		return err
	}

	switch vv := v.(type) {
	case string:
		switch {
		case strings.Index(vv, "https://") == 0 || strings.Index(vv, "http://") == 0:
			hc, err := newHTTPRunner(k, vv, nil)
			if err != nil {
				return err
			}
			bk.httpRunners[k] = hc
		case strings.Index(vv, "grpc://") == 0:
			addr := strings.TrimPrefix(vv, "grpc://")
			gc, err := newGrpcRunner(k, addr, nil)
			if err != nil {
				return err
			}
			bk.grpcRunners[k] = gc
		default:
			dc, err := newDBRunner(k, vv, nil)
			if err != nil {
				return err
			}
			bk.dbRunners[k] = dc
		}
	case map[string]interface{}:
		tmp, err := yaml.Marshal(vv)
		if err != nil {
			return err
		}
		detect := false

		// HTTP Runner
		c := &httpRunnerConfig{}
		if err := yaml.Unmarshal(tmp, c); err == nil {
			if c.Endpoint != "" {
				detect = true
				r, err := newHTTPRunner(k, c.Endpoint, nil)
				if err != nil {
					return err
				}
				if c.OpenApi3DocLocation != "" && !strings.HasPrefix(c.OpenApi3DocLocation, "https://") && !strings.HasPrefix(c.OpenApi3DocLocation, "http://") && !strings.HasPrefix(c.OpenApi3DocLocation, "/") {
					c.OpenApi3DocLocation = filepath.Join(root, c.OpenApi3DocLocation)
				}
				hv, err := newHttpValidator(c)
				if err != nil {
					return err
				}
				r.validator = hv
				bk.httpRunners[k] = r
			}
		}

		// gRPC Runner
		if !detect {
			c := &grpcRunnerConfig{}
			if err := yaml.Unmarshal(tmp, c); err == nil {
				if c.Addr != "" {
					detect = true
					r, err := newGrpcRunner(k, c.Addr, nil)
					if err != nil {
						return err
					}
					r.tls = c.TLS
					if c.cacert != nil {
						r.cacert = c.cacert
					} else if strings.HasPrefix(c.CACert, "/") {
						b, err := os.ReadFile(c.CACert)
						if err != nil {
							return err
						}
						r.cacert = b
					} else {
						b, err := os.ReadFile(filepath.Join(root, c.CACert))
						if err != nil {
							return err
						}
						r.cacert = b
					}
					if c.cert != nil {
						r.cert = c.cert
					} else if strings.HasPrefix(c.Cert, "/") {
						b, err := os.ReadFile(c.Cert)
						if err != nil {
							return err
						}
						r.cert = b
					} else {
						b, err := os.ReadFile(filepath.Join(root, c.Cert))
						if err != nil {
							return err
						}
						r.cert = b
					}
					if c.key != nil {
						r.key = c.key
					} else if strings.HasPrefix(c.Key, "/") {
						b, err := os.ReadFile(c.Key)
						if err != nil {
							return err
						}
						r.key = b
					} else {
						b, err := os.ReadFile(filepath.Join(root, c.Key))
						if err != nil {
							return err
						}
						r.key = b
					}
					r.skipVerify = c.SkipVerify
					bk.grpcRunners[k] = r
				}
			}
		}

		if !detect {
			return fmt.Errorf("cannot detect runner: %s", string(tmp))
		}
	}

	return nil
}

func validateRunnerKey(k string) error {
	if k == deprecatedRetrySectionKey {
		_, _ = fmt.Fprintf(os.Stderr, "'%s' is deprecated. use %s instead", deprecatedRetrySectionKey, loopSectionKey)
	}
	if k == includeRunnerKey || k == testRunnerKey || k == dumpRunnerKey || k == execRunnerKey || k == bindRunnerKey {
		return fmt.Errorf("runner name '%s' is reserved for built-in runner", k)
	}
	if k == ifSectionKey || k == descSectionKey || k == loopSectionKey || k == deprecatedRetrySectionKey {
		return fmt.Errorf("runner name '%s' is reserved for built-in section", k)
	}
	return nil
}

func LoadBook(path string) (*book, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}
	bk, err := loadBook(f)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}
	bk.path = path
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}

	return bk, nil
}

func (bk *book) applyOptions(opts ...Option) error {
	opts = setupBuiltinFunctions(opts...)
	for _, opt := range opts {
		if err := opt(bk); err != nil {
			return err
		}
	}
	return nil
}

func (bk *book) generateOperatorId() string {
	if bk.path != "" {
		return bk.path
	} else {
		return xid.New().String()
	}
}

func (bk *book) generateOperatorRoot() (string, error) {
	if bk.path != "" {
		return filepath.Dir(bk.path), nil
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return wd, nil
	}
}

func validateStepKeys(s map[string]interface{}) error {
	if len(s) == 0 {
		return errors.New("step must specify at least one runner")
	}
	custom := 0
	for k := range s {
		if k == testRunnerKey || k == dumpRunnerKey || k == bindRunnerKey || k == ifSectionKey || k == descSectionKey || k == loopSectionKey || k == deprecatedRetrySectionKey {
			continue
		}
		custom += 1
	}
	if custom > 1 {
		return errors.New("runners that cannot be running at the same time are specified")
	}
	return nil
}
