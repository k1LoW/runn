package runn

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const noDesc = "[No Description]"
const testRunnerKey = "test"

type book struct {
	Desc        string                   `yaml:"desc,omitempty"`
	Runners     map[string]string        `yaml:"runners,omitempty"`
	Vars        map[string]string        `yaml:"vars,omitempty"`
	Steps       []map[string]interface{} `yaml:"steps,omitempty"`
	Debug       bool                     `yaml:"debug,omitempty"`
	httpRunners map[string]*httpRunner
	dbRunners   map[string]*dbRunner
	t           *testing.T
}

func newBook() *book {
	return &book{
		Runners:     map[string]string{},
		Vars:        map[string]string{},
		Steps:       []map[string]interface{}{},
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
	}
}

func LoadBook(in io.Reader) (*book, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, in); err != nil {
		return nil, err
	}
	bk := newBook()
	if err := yaml.Unmarshal(expand.ExpandenvYAMLBytes(buf.Bytes()), bk); err != nil {
		return nil, err
	}
	if bk.Runners == nil {
		bk.Runners = map[string]string{}
	}
	if bk.Vars == nil {
		bk.Vars = map[string]string{}
	}
	if bk.Desc == "" {
		bk.Desc = noDesc
	}
	return bk, nil
}

func LoadBookFile(path string) (*book, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	bk, err := LoadBook(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	return bk, nil
}
