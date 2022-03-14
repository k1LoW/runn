package runn

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
)

const noDesc = "[No Description]"

type book struct {
	Desc        string                   `yaml:"desc,omitempty"`
	Runners     map[string]string        `yaml:"runners,omitempty"`
	Vars        map[string]interface{}   `yaml:"vars,omitempty"`
	Steps       []map[string]interface{} `yaml:"steps,omitempty"`
	Debug       bool                     `yaml:"debug,omitempty"`
	stepKeys    []string
	path        string
	httpRunners map[string]*httpRunner
	dbRunners   map[string]*dbRunner
	t           *testing.T
}

func newBook() *book {
	return &book{
		Runners:     map[string]string{},
		Vars:        map[string]interface{}{},
		Steps:       []map[string]interface{}{},
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
	}
}

func loadBook(in io.Reader) (*book, error) {
	bk := newBook()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, in); err != nil {
		return nil, err
	}
	if err := yaml.NewDecoder(bytes.NewBuffer(expand.ExpandenvYAMLBytes(buf.Bytes()))).Decode(bk); err == nil {
		if bk.Runners == nil {
			bk.Runners = map[string]string{}
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
		Desc    string                 `yaml:"desc,omitempty"`
		Runners map[string]string      `yaml:"runners,omitempty"`
		Vars    map[string]interface{} `yaml:"vars,omitempty"`
		Steps   yaml.MapSlice          `yaml:"steps,omitempty"`
		Debug   bool                   `yaml:"debug,omitempty"`
	}{
		Runners: map[string]string{},
		Vars:    map[string]interface{}{},
		Steps:   yaml.MapSlice{},
	}

	if err := yaml.NewDecoder(bytes.NewBuffer(expand.ExpandenvYAMLBytes(buf.Bytes()))).Decode(&m); err != nil {
		return nil, err
	}
	bk.Desc = m.Desc
	bk.Runners = m.Runners
	bk.Vars = m.Vars
	bk.Debug = m.Debug
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
