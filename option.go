package runn

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

type Option func(*book) error

// Book - Load runbook
func Book(path string) Option {
	return func(bk *book) error {
		loaded, err := LoadBook(path)
		if err != nil {
			return err
		}
		bk.Desc = loaded.Desc
		bk.If = loaded.If
		for k, r := range loaded.Runners {
			if r != nil {
				bk.Runners[k] = r
			}
		}
		for k, r := range loaded.httpRunners {
			if r != nil {
				bk.httpRunners[k] = r
			}
		}
		for k, r := range loaded.dbRunners {
			if r != nil {
				bk.dbRunners[k] = r
			}
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
func Runner(name, dsn string, opts ...RunnerOption) Option {
	return func(bk *book) error {
		delete(bk.runnerErrs, name)
		if len(opts) == 0 {
			bk.Runners[name] = dsn
			return nil
		}
		c := &RunnerConfig{}
		for _, opt := range opts {
			if err := opt(c); err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
		}
		switch {
		case c.OpenApi3DocLocation != "":
			r, err := newHTTPRunner(name, dsn, nil)
			if err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
			v, err := NewHttpValidator(c)
			if err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
			r.validator = v
			bk.httpRunners[name] = r
		default:
			bk.runnerErrs[name] = errors.New("invalid runner option")
			return nil
		}
		return nil
	}
}

// HTTPRunner - Set http runner to runbook
func HTTPRunner(name, endpoint string, client *http.Client, opts ...RunnerOption) Option {
	return func(bk *book) error {
		delete(bk.runnerErrs, name)
		r, err := newHTTPRunner(name, endpoint, nil)
		if err != nil {
			return err
		}
		r.client = client
		bk.httpRunners[name] = r
		if len(opts) == 0 {
			return nil
		}
		c := &RunnerConfig{}
		for _, opt := range opts {
			if err := opt(c); err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
		}
		v, err := NewHttpValidator(c)
		if err != nil {
			bk.runnerErrs[name] = err
			return nil
		}
		r.validator = v
		return nil
	}
}

// HTTPRunner - Set http runner to runbook with http.Handler
func HTTPRunnerWithHandler(name string, h http.Handler, opts ...RunnerOption) Option {
	return func(bk *book) error {
		delete(bk.runnerErrs, name)
		r, err := newHTTPRunnerWithHandler(name, h, nil)
		if err != nil {
			bk.runnerErrs[name] = err
			return nil
		}
		if len(opts) > 0 {
			c := &RunnerConfig{}
			for _, opt := range opts {
				if err := opt(c); err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
			}
			v, err := NewHttpValidator(c)
			if err != nil {
				bk.runnerErrs[name] = err
				return nil
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
		delete(bk.runnerErrs, name)
		bk.dbRunners[name] = &dbRunner{
			name:   name,
			client: client,
		}
		return nil
	}
}

// T - Acts as test helper
func T(t *testing.T) Option {
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

// Func - Set function to runner
func Func(k string, v interface{}) Option {
	return func(bk *book) error {
		bk.Funcs[k] = v
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

// SkipIncluded - Skip running the included step by itself.
func SkipIncluded(enable bool) Option {
	return func(bk *book) error {
		bk.skipIncluded = enable
		return nil
	}
}

func included(included bool) Option {
	return func(bk *book) error {
		bk.included = included
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

func runnHTTPRunner(name string, r *httpRunner) Option {
	return func(bk *book) error {
		bk.httpRunners[name] = r
		return nil
	}
}

func runnDBRunner(name string, r *dbRunner) Option {
	return func(bk *book) error {
		bk.dbRunners[name] = r
		return nil
	}
}

var (
	AsTestHelper = T
	Runbook      = Book
)
