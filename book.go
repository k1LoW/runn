package runn

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/duration"
	"github.com/k1LoW/expand"
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

func loadBook(in io.Reader) (*book, error) {
	bk := newBook()
	b, err := io.ReadAll(in)
	if err != nil {
		return nil, err
	}
	b = expand.ExpandenvYAMLBytes(b)
	if err := yaml.Unmarshal(b, bk); err == nil {
		if bk.Runners == nil {
			bk.Runners = map[string]interface{}{}
		}
		if bk.Vars == nil {
			bk.Vars = map[string]interface{}{}
		}
		if bk.Desc == "" {
			bk.Desc = noDesc
		}
		return bk, nil
	}

	// orderedmap
	m := struct {
		Desc     string                 `yaml:"desc,omitempty"`
		Runners  map[string]interface{} `yaml:"runners,omitempty"`
		Vars     map[string]interface{} `yaml:"vars,omitempty"`
		Steps    yaml.MapSlice          `yaml:"steps,omitempty"`
		Debug    bool                   `yaml:"debug,omitempty"`
		Interval string                 `yaml:"interval,omitempty"`
		If       string                 `yaml:"if,omitempty"`
		SkipTest bool                   `yaml:"skipTest,omitempty"`
	}{
		Runners: map[string]interface{}{},
		Vars:    map[string]interface{}{},
		Steps:   yaml.MapSlice{},
	}

	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
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
			return nil, err
		}
		bk.interval = d
	}

	for _, s := range m.Steps {
		bk.Steps = append(bk.Steps, s.Value.(map[string]interface{}))
		switch v := s.Key.(type) {
		case string:
			bk.stepKeys = append(bk.stepKeys, v)
		case uint64:
			bk.stepKeys = append(bk.stepKeys, fmt.Sprintf("%d", v))
		default:
			bk.stepKeys = append(bk.stepKeys, fmt.Sprintf("%v", v))
		}
	}
	return bk, nil
}

func LoadBook(path string) (*book, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	bk, err := loadBook(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	bk.path = path
	if err := f.Close(); err != nil {
		return nil, err
	}

	return bk, nil
}
