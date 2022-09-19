package runn

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/k1LoW/duration"
	"github.com/k1LoW/expand"
	"gopkg.in/yaml.v2"
)

const noDesc = "[No Description]"

type book struct {
	Desc          string                   `yaml:"desc,omitempty"`
	Runners       map[string]interface{}   `yaml:"runners,omitempty"`
	Vars          map[string]interface{}   `yaml:"vars,omitempty"`
	Funcs         map[string]interface{}   `yaml:"-"`
	Steps         []map[string]interface{} `yaml:"steps,omitempty"`
	Debug         bool                     `yaml:"debug,omitempty"`
	Interval      string                   `yaml:"interval,omitempty"`
	If            string                   `yaml:"if,omitempty"`
	SkipTest      bool                     `yaml:"skipTest,omitempty"`
	stepKeys      []string
	path          string // runbook file path
	httpRunners   map[string]*httpRunner
	dbRunners     map[string]*dbRunner
	grpcRunners   map[string]*grpcRunner
	profile       bool
	interval      time.Duration
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
		Funcs:       map[string]interface{}{},
		Steps:       []map[string]interface{}{},
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
	if err := yamlUnmarshal(b, bk); err == nil {
		if bk.Runners == nil {
			bk.Runners = map[string]interface{}{}
		} else {
			bk.Runners = normalize(bk.Runners).(map[string]interface{})
		}
		if bk.Vars == nil {
			bk.Vars = map[string]interface{}{}
		} else {
			bk.Vars = normalize(bk.Vars).(map[string]interface{})
			// To match behavior with json.Marshal
			b, err := json.Marshal(bk.Vars)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(b, &bk.Vars); err != nil {
				return nil, err
			}
		}
		bk.Steps = normalize(bk.Steps).([]map[string]interface{})
		if bk.Desc == "" {
			bk.Desc = noDesc
		}
		return bk, nil
	}

	// orderedmap
	m := newMapped()
	if err := yamlUnmarshal(b, &m); err != nil {
		return nil, err
	}
	bk.Desc = m.Desc
	bk.Runners = normalize(m.Runners).(map[string]interface{})
	bk.Vars = normalize(m.Vars).(map[string]interface{})
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
			return nil, err
		}
		bk.interval = d
	}

	keys := map[string]struct{}{}
	for _, s := range m.Steps {
		bk.Steps = append(bk.Steps, normalize(s.Value).(map[string]interface{}))
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
			return nil, fmt.Errorf("duplicate step keys: %s", k)
		}
		keys[k] = struct{}{}
	}
	return bk, nil
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
