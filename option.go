package runbk

import (
	"database/sql"
	"net/http"
	"net/url"
	"testing"
)

type Option func(*book) error

func Book(path string) Option {
	return func(bk *book) error {
		loaded, err := LoadBookFile(path)
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

func Desc(desc string) Option {
	return func(bk *book) error {
		bk.Desc = desc
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

func AsTestHelper(t *testing.T) Option {
	return func(bk *book) error {
		bk.t = t
		return nil
	}
}

var T = AsTestHelper
