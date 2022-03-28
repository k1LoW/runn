package runn

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

type Option func(*book) error

// Book - load runbook
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
		for k, r := range loaded.httpRunners {
			bk.httpRunners[k] = r
		}
		for k, r := range loaded.dbRunners {
			bk.dbRunners[k] = r
		}
		for k, v := range loaded.Vars {
			bk.Vars[k] = v
		}
		bk.Steps = loaded.Steps
		bk.stepKeys = loaded.stepKeys
		if !bk.Debug {
			bk.Debug = loaded.Debug
		}
		if loaded.Interval != "" {
			bk.Interval = loaded.Interval
			bk.interval = loaded.interval
		}
		bk.path = loaded.path
		return nil
	}
}

// Desc - Set description to runbook
func Desc(desc string) Option {
	return func(bk *book) error {
		bk.Desc = desc
		return nil
	}
}

// Runner - Set runner to runbook
func Runner(name, dsn string) Option {
	return func(bk *book) error {
		bk.Runners[name] = dsn
		return nil
	}
}

// HTTPRunner - Set http runner to runbook
func HTTPRunner(name, endpoint string, client *http.Client, opts ...RunnerOption) Option {
	return func(bk *book) error {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		r := &httpRunner{
			name:     name,
			endpoint: u,
			client:   client,
		}
		if len(opts) > 0 {
			c := &RunnerConfig{}
			for _, opt := range opts {
				if err := opt(c); err != nil {
					return err
				}
			}
			v, err := NewHttpValidator(c)
			if err != nil {
				return err
			}
			r.validator = v
		}
		bk.httpRunners[name] = r
		return nil
	}
}

// HTTPRunner - Set http runner to runbook with http.Handler
func HTTPRunnerWithHandler(name string, h http.Handler, opts ...RunnerOption) Option {
	return func(bk *book) error {
		r := &httpRunner{
			name:    name,
			handler: h,
		}
		if len(opts) > 0 {
			c := &RunnerConfig{}
			for _, opt := range opts {
				if err := opt(c); err != nil {
					return err
				}
			}
			v, err := NewHttpValidator(c)
			if err != nil {
				return err
			}
			r.validator = v
		}
		bk.httpRunners[name] = r
		return nil
	}
}

// DBRunner - Set db runner to runbook
func DBRunner(name string, client *sql.DB) Option {
	return func(bk *book) error {
		bk.dbRunners[name] = &dbRunner{
			name:   name,
			client: client,
		}
		return nil
	}
}

// AsTestHelper - Acts as test helper
func AsTestHelper(t *testing.T) Option {
	return func(bk *book) error {
		bk.t = t
		return nil
	}
}

// Var - Set variable to runner
func Var(k string, v interface{}) Option {
	return func(bk *book) error {
		bk.Vars[k] = v
		return nil
	}
}

// Debug - Enable debug output
func Debug(debug bool) Option {
	return func(bk *book) error {
		if !bk.Debug {
			bk.Debug = debug
		}
		return nil
	}
}

// Interval - Set interval between steps
func Interval(d time.Duration) Option {
	return func(bk *book) error {
		if d < 0 {
			return fmt.Errorf("invalid interval: %s", d)
		}
		bk.interval = d
		return nil
	}
}

// FailFast - Enable fail-fast
func FailFast(enable bool) Option {
	return func(bk *book) error {
		bk.failFast = enable
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
