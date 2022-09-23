package runn

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/araddon/dateparse"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jhump/protoreflect/desc"
	"github.com/spf13/cast"
	"google.golang.org/grpc"
)

type Option func(*book) error

// Book - Load runbook
func Book(path string) Option {
	return func(bk *book) error {
		loaded, err := LoadBook(path)
		if err != nil {
			return err
		}
		bk.desc = loaded.desc
		bk.ifCond = loaded.ifCond
		bk.useMap = loaded.useMap
		for k, r := range loaded.runners {
			if r != nil {
				bk.runners[k] = r
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
		for k, r := range loaded.grpcRunners {
			if r != nil {
				bk.grpcRunners[k] = r
			}
		}
		for k, v := range loaded.vars {
			root, err := loaded.generateOperatorRoot()
			if err != nil {
				return err
			}
			ev, err := evaluateSchema(v, root, nil)
			if err != nil {
				return err
			}
			bk.vars[k] = ev
		}
		bk.rawSteps = loaded.rawSteps
		bk.stepKeys = loaded.stepKeys
		if !bk.debug {
			bk.debug = loaded.debug
		}
		if !bk.skipTest {
			bk.skipTest = loaded.skipTest
		}
		if loaded.intervalStr != "" {
			bk.interval = loaded.interval
		}
		bk.path = loaded.path
		return nil
	}
}

// Desc - Set description to runbook
func Desc(desc string) Option {
	return func(bk *book) error {
		bk.desc = desc
		return nil
	}
}

// Runner - Set runner to runbook
func Runner(name, dsn string, opts ...httpRunnerOption) Option {
	return func(bk *book) error {
		delete(bk.runnerErrs, name)
		if len(opts) == 0 {
			if err := validateRunnerKey(name); err != nil {
				return err
			}
			if err := bk.parseRunner(name, dsn); err != nil {
				bk.runnerErrs[name] = err
			}
			return nil
		}
		c := &httpRunnerConfig{}
		for _, opt := range opts {
			if err := opt(c); err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
		}
		switch {
		case c.OpenApi3DocLocation != "":
			r, err := newHTTPRunner(name, dsn)
			if err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
			v, err := newHttpValidator(c)
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
func HTTPRunner(name, endpoint string, client *http.Client, opts ...httpRunnerOption) Option {
	return func(bk *book) error {
		delete(bk.runnerErrs, name)
		r, err := newHTTPRunner(name, endpoint)
		if err != nil {
			return err
		}
		r.client = client
		bk.httpRunners[name] = r
		if len(opts) == 0 {
			return nil
		}
		c := &httpRunnerConfig{}
		for _, opt := range opts {
			if err := opt(c); err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
		}
		v, err := newHttpValidator(c)
		if err != nil {
			bk.runnerErrs[name] = err
			return nil
		}
		r.validator = v
		return nil
	}
}

// HTTPRunnerWithHandler - Set http runner to runbook with http.Handler
func HTTPRunnerWithHandler(name string, h http.Handler, opts ...httpRunnerOption) Option {
	return func(bk *book) error {
		delete(bk.runnerErrs, name)
		r, err := newHTTPRunnerWithHandler(name, h)
		if err != nil {
			bk.runnerErrs[name] = err
			return nil
		}
		if len(opts) > 0 {
			c := &httpRunnerConfig{}
			for _, opt := range opts {
				if err := opt(c); err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
			}
			v, err := newHttpValidator(c)
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

// GrpcRunner - Set grpc runner to runbook
func GrpcRunner(name string, cc *grpc.ClientConn, opts ...grpcRunnerOption) Option {
	return func(bk *book) error {
		delete(bk.runnerErrs, name)
		r := &grpcRunner{
			name: name,
			cc:   cc,
			mds:  map[string]*desc.MethodDescriptor{},
		}
		if len(opts) > 0 {
			c := &grpcRunnerConfig{}
			for _, opt := range opts {
				if err := opt(c); err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
			}
			r.tls = c.TLS
			if c.cacert != nil {
				r.cacert = c.cacert
			} else {
				b, err := os.ReadFile(c.CACert)
				if err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
				r.cacert = b
			}
			if c.cert != nil {
				r.cert = c.cert
			} else {
				b, err := os.ReadFile(c.Cert)
				if err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
				r.cert = b
			}
			if c.key != nil {
				r.key = c.key
			} else {
				b, err := os.ReadFile(c.Key)
				if err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
				r.key = b
			}
			r.skipVerify = c.SkipVerify
		}
		bk.grpcRunners[name] = r
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
		root, err := bk.generateOperatorRoot()
		if err != nil {
			return err
		}
		ev, err := evaluateSchema(v, root, nil)
		if err != nil {
			return err
		}
		bk.vars[k] = ev
		return nil
	}
}

// Func - Set function to runner
func Func(k string, v interface{}) Option {
	return func(bk *book) error {
		bk.funcs[k] = v
		return nil
	}
}

// Debug - Enable debug output
func Debug(debug bool) Option {
	return func(bk *book) error {
		if !bk.debug {
			bk.debug = debug
		}
		return nil
	}
}

// Profile - Enable profile output
func Profile(profile bool) Option {
	return func(bk *book) error {
		if !bk.profile {
			bk.profile = profile
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

// SkipTest - Skip test section
func SkipTest(enable bool) Option {
	return func(bk *book) error {
		if !bk.skipTest {
			bk.skipTest = enable
		}
		return nil
	}
}

// BeforeFunc - Register the function to be run before the runbook is run.
func BeforeFunc(fn func() error) Option {
	return func(bk *book) error {
		bk.beforeFuncs = append(bk.beforeFuncs, fn)
		return nil
	}
}

// AfterFunc - Register the function to be run after the runbook is run.
func AfterFunc(fn func() error) Option {
	return func(bk *book) error {
		bk.afterFuncs = append(bk.afterFuncs, fn)
		return nil
	}
}

// Capture - Register the capturer to capture steps.
func Capture(c Capturer) Option {
	return func(bk *book) error {
		bk.capturers = append(bk.capturers, c)
		return nil
	}
}

// setupBuiltinFunctions - Set up built-in functions to runner
func setupBuiltinFunctions(opts ...Option) []Option {
	// Built-in functions are added at the beginning of an option and are overridden by subsequent options
	return append([]Option{
		// NOTE: Please add here the built-in functions you want to enable.
		Func("urlencode", url.QueryEscape),
		Func("string", func(v interface{}) string { return cast.ToString(v) }),
		Func("int", func(v interface{}) int { return cast.ToInt(v) }),
		Func("bool", func(v interface{}) bool { return cast.ToBool(v) }),
		Func("time", func(v interface{}) time.Time {
			t, err := dateparse.ParseStrict(v.(string))
			if err != nil {
				return time.Time{}
			}
			return t
		}),
		Func("compare", func(x, y interface{}, ignoreKeys ...string) bool {
			// normalize values
			bx, err := json.Marshal(x)
			if err != nil {
				return false
			}
			var vx interface{}
			if err := json.Unmarshal(bx, &vx); err != nil {
				return false
			}
			by, err := json.Marshal(x)
			if err != nil {
				return false
			}
			var vy interface{}
			if err := json.Unmarshal(by, &vy); err != nil {
				return false
			}

			diff := cmp.Diff(vx, vy, cmpopts.IgnoreMapEntries(func(key string, val interface{}) bool {
				for _, ignore := range ignoreKeys {
					if key == ignore {
						return true
					}
				}
				return false
			}))

			// FIXME: Debug output of diffs

			return diff == ""

		}),
	},
		opts...,
	)
}

// RunMatch - Run only runbooks with matching paths.
func RunMatch(m string) Option {
	return func(bk *book) error {
		re, err := regexp.Compile(m)
		if err != nil {
			return err
		}
		bk.runMatch = re
		return nil
	}
}

// RunSample - Run the specified number of runbooks at random.
func RunSample(n int) Option {
	return func(bk *book) error {
		if n <= 0 {
			return fmt.Errorf("sample must be greater than 0: %d", n)
		}
		bk.runSample = n
		return nil
	}
}

// RunShard - Distribute runbooks into a specified number of shards and run the specified shard of them.
func RunShard(n, i int) Option {
	return func(bk *book) error {
		if n <= 0 {
			return fmt.Errorf("the number of divisions is greater than 0: %d", n)
		}
		if i < 0 {
			return fmt.Errorf("the index of divisions is greater than or equal to 0: %d", i)
		}
		if i >= n {
			return fmt.Errorf("the index of divisions is less than the number of distributions (%d): %d", n, i)
		}
		bk.runShardIndex = i
		bk.runShardN = n
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

func GetDesc(opt Option) (string, error) {
	b := newBook()
	if err := opt(b); err != nil {
		return "", err
	}
	return b.desc, nil
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

func runnGrpcRunner(name string, r *grpcRunner) Option {
	return func(bk *book) error {
		bk.grpcRunners[name] = r
		return nil
	}
}

var (
	AsTestHelper = T
	Runbook      = Book
	RunPart      = RunShard
)
