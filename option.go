package runn

import (
	"database/sql"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
)

type Option func(*book) error

func Book(path string) Option {
	return func(bk *book) error {
		loaded, err := LoadBook(path)
		if err != nil {
			return err
		}
		bk.Desc = loaded.Desc
		for k, r := range loaded.Runners {
			bk.Runners[k] = r
		}
		for k, v := range loaded.Vars {
			bk.Vars[k] = v
		}
		bk.Steps = loaded.Steps
		bk.Debug = loaded.Debug
		bk.path = loaded.path
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
			name:     name,
			endpoint: u,
			client:   client,
		}
		return nil
	}
}

func DBRunner(name string, client *sql.DB) Option {
	return func(bk *book) error {
		bk.dbRunners[name] = &dbRunner{
			name:   name,
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

func Var(k string, v interface{}) Option {
	return func(bk *book) error {
		bk.Vars[k] = v
		return nil
	}
}

func Debug(debug bool) Option {
	return func(bk *book) error {
		if !bk.Debug {
			bk.Debug = debug
		}
		return nil
	}
}

func Books(pathp string) ([]Option, error) {
	paths, err := Paths(pathp)
	if err != nil {
		return nil, err
	}
	opts := []Option{}
	for _, p := range paths {
		opts = append(opts, Book(p))
	}
	return opts, nil
}

func Paths(pathp string) ([]string, error) {
	paths := []string{}
	base, pattern := doublestar.SplitPattern(pathp)
	abs, err := filepath.Abs(base)
	if err != nil {
		return nil, err
	}
	fsys := os.DirFS(abs)
	if err := doublestar.GlobWalk(fsys, pattern, func(p string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		paths = append(paths, filepath.Join(base, p))
		return nil
	}); err != nil {
		return nil, err
	}
	return paths, nil
}

func GetDesc(opt Option) string {
	b := newBook()
	_ = opt(b)
	return b.Desc
}

var (
	T       = AsTestHelper
	Runbook = Book
)
