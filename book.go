package runbk

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

const testRunnerKey = "test"

type book struct {
	Desc        string                   `yaml:"desc,omitempty"`
	Runners     map[string]string        `yaml:"runners,omitempty"`
	Vars        map[string]string        `yaml:"vars,omitempty"`
	Steps       []map[string]interface{} `yaml:"steps,omitempty"`
	httpRunners map[string]*httpRunner
	dbRunners   map[string]*dbRunner
}

type Option func(*book) error

func Book(path string) Option {
	return func(bk *book) error {
		loaded, err := loadBookFile(path)
		if err != nil {
			return err
		}
		bk.Desc = loaded.Desc
		bk.Runners = loaded.Runners
		bk.Vars = loaded.Vars
		bk.Steps = loaded.Steps
		return nil
	}
}

func Runner(name, dsn string) Option {
	return func(bk *book) error {
		bk.Runners[name] = dsn
		return nil
	}
}

func HTTPRunner(name, endpoint string, client *http.Client) Option {
	return func(bk *book) error {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		bk.httpRunners[name] = &httpRunner{
			endpoint: u,
			client:   client,
		}
		return nil
	}
}

func DBRunner(name string, client *sql.DB) Option {
	return func(bk *book) error {
		bk.dbRunners[name] = &dbRunner{
			client: client,
		}
		return nil
	}
}

func loadBook(in io.Reader) (*book, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, in); err != nil {
		return nil, err
	}
	bk := &book{
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
	}
	if err := yaml.Unmarshal(expand.ExpandenvYAMLBytes(buf.Bytes()), bk); err != nil {
		return nil, err
	}
	if bk.Runners == nil {
		bk.Runners = map[string]string{}
	}
	if bk.Vars == nil {
		bk.Vars = map[string]string{}
	}
	return bk, nil
}

func loadBookFile(path string) (*book, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	bk, err := loadBook(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	return bk, nil
}
