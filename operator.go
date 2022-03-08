package runn

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/antonmedv/expr"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
)

var expandRe = regexp.MustCompile(`"?{{\s*([^}]+)\s*}}"?`)

type step struct {
	httpRunner  *httpRunner
	httpRequest map[string]interface{}
	dbRunner    *dbRunner
	dbQuery     map[string]interface{}
	testRunner  *testRunner
	testCond    string
	debug       bool
}

type store struct {
	steps []map[string]interface{}
	vars  map[string]string
}

type operator struct {
	httpRunners map[string]*httpRunner
	dbRunners   map[string]*dbRunner
	steps       []*step
	store       store
	desc        string
	debug       bool
	t           *testing.T
}

func New(opts ...Option) (*operator, error) {
	bk := newBook()
	for _, opt := range opts {
		if err := opt(bk); err != nil {
			return nil, err
		}
	}
	o := &operator{
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
		store: store{
			steps: []map[string]interface{}{},
			vars:  bk.Vars,
		},
		desc:  bk.Desc,
		debug: bk.Debug,
		t:     bk.t,
	}

	for k, v := range bk.Runners {
		switch {
		case strings.Index(v, "https://") == 0 || strings.Index(v, "http://") == 0:
			hc, err := newHTTPRunner(k, v, o)
			if err != nil {
				return nil, err
			}
			o.httpRunners[k] = hc
		default:
			dc, err := newDBRunner(k, v, o)
			if err != nil {
				return nil, err
			}
			o.dbRunners[k] = dc
		}
	}
	for k, v := range bk.httpRunners {
		v.operator = o
		o.httpRunners[k] = v
	}
	for k, v := range bk.dbRunners {
		v.operator = o
		o.dbRunners[k] = v
	}

	for i, s := range bk.Steps {
		if len(s) != 1 {
			return nil, fmt.Errorf("invalid steps[%d]: %v", i, s)
		}
		if err := o.AppendStep(s); err != nil {
			return nil, err
		}
	}

	return o, nil
}

func (o *operator) AppendStep(s map[string]interface{}) error {
	if o.t != nil {
		o.t.Helper()
	}
	step := &step{debug: o.debug}
	for k, v := range s {
		if k == testRunnerKey {
			tr, err := newTestRunner(o)
			if err != nil {
				return err
			}
			step.testRunner = tr
			vv, ok := v.(string)
			if !ok {
				return fmt.Errorf("invalid test condition: %v", v)
			}
			step.testCond = vv
			continue
		}
		h, ok := o.httpRunners[k]
		if ok {
			step.httpRunner = h
			vv, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid http request: %v", v)
			}
			step.httpRequest = vv
			continue
		}
		db, ok := o.dbRunners[k]
		if ok {
			step.dbRunner = db
			vv, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid http request: %v", v)
			}
			step.dbQuery = vv
			continue
		}
		return fmt.Errorf("can not find client: %s", k)
	}
	o.steps = append(o.steps, step)
	return nil
}

func (o *operator) Run(ctx context.Context) error {
	if o.t != nil {
		o.t.Helper()
		o.t.Run(o.desc, func(t *testing.T) {
			t.Helper()
			if err := o.run(ctx); err != nil {
				t.Error(err)
			}
		})
		return nil
	}
	return o.run(ctx)
}

func (o *operator) run(ctx context.Context) error {
	for i, s := range o.steps {
		switch {
		case s.httpRunner != nil && s.httpRequest != nil:
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on steps[%d]\n", s.httpRunner.name, i)
			}
			e, err := o.expand(s.httpRequest)
			if err != nil {
				return err
			}
			r, ok := e.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid steps[%d]: %v", i, e)
			}
			req, err := parseHTTPRequest(r)
			if err != nil {
				return err
			}
			if err := s.httpRunner.Run(ctx, req); err != nil {
				return fmt.Errorf("http request failed on steps[%d]: %v", i, err)
			}
		case s.dbRunner != nil && s.dbQuery != nil:
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on steps[%d]\n", s.dbRunner.name, i)
			}
			e, err := o.expand(s.dbQuery)
			if err != nil {
				return err
			}
			q, ok := e.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid steps[%d]: %v", i, e)
			}
			query, err := parseDBQuery(q)
			if err != nil {
				return fmt.Errorf("invalid steps[%d]: %v", i, q)
			}
			if err := s.dbRunner.Run(ctx, query); err != nil {
				return fmt.Errorf("db query failed on steps[%d]: %v", i, err)
			}
		case s.testRunner != nil && s.testCond != "":
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on steps[%d]\n", testRunnerKey, i)
			}
			if err := s.testRunner.Run(ctx, s.testCond); err != nil {
				return fmt.Errorf("test failed on steps[%d]: %s", i, s.testCond)
			}
		default:
			return fmt.Errorf("invalid steps[%d]: %v", i, s)
		}
	}
	return nil
}

func (o *operator) expand(in interface{}) (interface{}, error) {
	store := map[string]interface{}{
		"steps": o.store.steps,
		"vars":  o.store.vars,
	}
	b, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	var reperr error
	replacefunc := func(in string) string {
		if !strings.Contains(in, "{{") {
			return in
		}
		matches := expandRe.FindAllStringSubmatch(in, -1)
		oldnew := []string{}
		for _, m := range matches {
			o, err := expr.Eval(m[1], store)
			if err != nil {
				reperr = err
			}
			var s string
			switch v := o.(type) {
			case string:
				s = v
			case int64:
				s = strconv.Itoa(int(v))
			default:
				reperr = fmt.Errorf("invalid expand format: %v", o)
			}
			oldnew = append(oldnew, m[0], s)
		}
		rep := strings.NewReplacer(oldnew...)
		return rep.Replace(in)
	}
	e := expand.ReplaceYAML(string(b), replacefunc, true)
	if reperr != nil {
		return nil, reperr
	}
	var out interface{}
	if err := yaml.Unmarshal([]byte(e), &out); err != nil {
		return nil, err
	}
	return out, nil
}

type operators []*operator

func Load(path string, opts ...Option) (operators, error) {
	base, pattern := doublestar.SplitPattern(path)
	abs, err := filepath.Abs(base)
	if err != nil {
		return nil, err
	}
	ops := operators{}
	fsys := os.DirFS(abs)
	if err := doublestar.GlobWalk(fsys, pattern, func(p string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		o, err := New(append(opts, Book(filepath.Join(base, p)))...)
		if err != nil {
			return err
		}
		ops = append(ops, o)
		return nil
	}); err != nil {
		return nil, err
	}
	return ops, nil
}

func (ops operators) RunN(ctx context.Context) error {
	for _, o := range ops {
		if err := o.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
