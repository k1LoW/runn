package runn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/antonmedv/expr"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
)

var expandRe = regexp.MustCompile(`"?{{\s*([^}]+)\s*}}"?`)
var numberRe = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?$`)

type step struct {
	key           string
	httpRunner    *httpRunner
	httpRequest   map[string]interface{}
	dbRunner      *dbRunner
	dbQuery       map[string]interface{}
	execRunner    *execRunner
	execCommand   map[string]interface{}
	testRunner    *testRunner
	testCond      string
	dumpRunner    *dumpRunner
	dumpCond      string
	includeRunner *includeRunner
	includePath   string
	debug         bool
}

type store struct {
	steps    []map[string]interface{}
	stepMaps map[string]interface{}
	vars     map[string]interface{}
	useMaps  bool
}

func (s *store) toMap() map[string]interface{} {
	store := map[string]interface{}{
		"vars": s.vars,
	}
	if s.useMaps {
		store["steps"] = s.stepMaps
	} else {
		store["steps"] = s.steps
	}
	return store
}

type operator struct {
	httpRunners map[string]*httpRunner
	dbRunners   map[string]*dbRunner
	steps       []*step
	store       store
	desc        string
	useMaps     bool
	debug       bool
	interval    time.Duration
	root        string
	t           *testing.T
}

func (o *operator) record(v map[string]interface{}) {
	if o.useMaps && len(o.steps) > 0 {
		o.store.stepMaps[o.steps[len(o.store.stepMaps)].key] = v
		return
	}
	o.store.steps = append(o.store.steps, v)
}

func New(opts ...Option) (*operator, error) {
	bk := newBook()
	for _, opt := range opts {
		if err := opt(bk); err != nil {
			return nil, err
		}
	}

	useMaps := false
	if len(bk.stepKeys) == len(bk.Steps) {
		useMaps = true
	}

	o := &operator{
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
		store: store{
			steps:    []map[string]interface{}{},
			stepMaps: map[string]interface{}{},
			vars:     bk.Vars,
			useMaps:  useMaps,
		},
		useMaps:  useMaps,
		desc:     bk.Desc,
		debug:    bk.Debug,
		interval: bk.interval,
		t:        bk.t,
	}

	if bk.path != "" {
		o.root = filepath.Dir(bk.path)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		o.root = wd
	}

	for k, v := range bk.Runners {
		switch {
		case k == includeRunnerKey || k == testRunnerKey || k == dumpRunnerKey || k == execRunnerKey:
			return nil, fmt.Errorf("runner name '%s' is reserved for built-in runner", k)
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

	keys := map[string]struct{}{}
	for k := range o.httpRunners {
		keys[k] = struct{}{}
	}
	for k := range o.dbRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names: %s", k)
		}
		keys[k] = struct{}{}
	}

	for i, s := range bk.Steps {
		if len(s) != 1 {
			return nil, fmt.Errorf("invalid steps[%d]: %v", i, s)
		}
		key := ""
		if len(bk.stepKeys) == len(bk.Steps) {
			key = bk.stepKeys[i]
		}
		if err := o.AppendStep(key, s); err != nil {
			return nil, err
		}
	}

	return o, nil
}

func (o *operator) AppendStep(key string, s map[string]interface{}) error {
	if o.t != nil {
		o.t.Helper()
	}
	step := &step{key: key, debug: o.debug}
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
		if k == dumpRunnerKey {
			dr, err := newDumpRunner(o)
			if err != nil {
				return err
			}
			step.dumpRunner = dr
			vv, ok := v.(string)
			if !ok {
				return fmt.Errorf("invalid dump condition: %v", v)
			}
			step.dumpCond = vv
			continue
		}
		if k == includeRunnerKey {
			ir, err := newIncludeRunner(o)
			if err != nil {
				return err
			}
			step.includeRunner = ir
			vv, ok := v.(string)
			if !ok {
				return fmt.Errorf("invalid include path: %v", v)
			}
			step.includePath = vv
			continue
		}
		if k == execRunnerKey {
			er, err := newExecRunner(o)
			if err != nil {
				return err
			}
			step.execRunner = er
			vv, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid exec command: %v", v)
			}
			step.execCommand = vv
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
		var err error
		o.t.Run(o.desc, func(t *testing.T) {
			t.Helper()
			err = o.run(ctx)
			if err != nil {
				t.Error(err)
			}
		})
		return err
	}
	return o.run(ctx)
}

func (o *operator) run(ctx context.Context) error {
	for i, s := range o.steps {
		if i != 0 {
			time.Sleep(o.interval)
		}
		switch {
		case s.httpRunner != nil && s.httpRequest != nil:
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on %s\n", s.httpRunner.name, o.stepName(i))
			}
			e, err := o.expand(s.httpRequest)
			if err != nil {
				return err
			}
			r, ok := e.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid %s: %v", o.stepName(i), e)
			}
			req, err := parseHTTPRequest(r)
			if err != nil {
				return err
			}
			if err := s.httpRunner.Run(ctx, req); err != nil {
				return fmt.Errorf("http request failed on %s: %v", o.stepName(i), err)
			}
		case s.dbRunner != nil && s.dbQuery != nil:
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on %s\n", s.dbRunner.name, o.stepName(i))
			}
			e, err := o.expand(s.dbQuery)
			if err != nil {
				return err
			}
			q, ok := e.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid %s: %v", o.stepName(i), e)
			}
			query, err := parseDBQuery(q)
			if err != nil {
				return fmt.Errorf("invalid %s: %v", o.stepName(i), q)
			}
			if err := s.dbRunner.Run(ctx, query); err != nil {
				return fmt.Errorf("db query failed on %s: %v", o.stepName(i), err)
			}
		case s.execRunner != nil && s.execCommand != nil:
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on %s\n", execRunnerKey, o.stepName(i))
			}
			e, err := o.expand(s.execCommand)
			if err != nil {
				return err
			}
			cmd, ok := e.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid %s: %v", o.stepName(i), e)
			}
			command, err := parseExecCommand(cmd)
			if err != nil {
				return fmt.Errorf("invalid %s: %v", o.stepName(i), cmd)
			}
			if err := s.execRunner.Run(ctx, command); err != nil {
				return fmt.Errorf("exec command failed on %s: %v", o.stepName(i), err)
			}
		case s.testRunner != nil && s.testCond != "":
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on %s\n", testRunnerKey, o.stepName(i))
			}
			if err := s.testRunner.Run(ctx, s.testCond); err != nil {
				return fmt.Errorf("test failed on %s: %v", o.stepName(i), err)
			}
		case s.dumpRunner != nil && s.dumpCond != "":
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on %s\n", dumpRunnerKey, o.stepName(i))
			}
			if err := s.dumpRunner.Run(ctx, s.dumpCond); err != nil {
				return fmt.Errorf("dump failed on %s: %v", o.stepName(i), err)
			}
		case s.includeRunner != nil && s.includePath != "":
			if o.debug {
				_, _ = fmt.Fprintf(os.Stderr, "Run '%s' on %s\n", includeRunnerKey, o.stepName(i))
			}
			if err := s.includeRunner.Run(ctx, s.includePath); err != nil {
				return fmt.Errorf("include failed on %s: %v", o.stepName(i), err)
			}
		default:
			return fmt.Errorf("invalid %s: %v", o.stepName(i), s)
		}
	}
	return nil
}

func (o *operator) stepName(i int) string {
	if o.useMaps {
		return fmt.Sprintf("%s.steps.%s", o.desc, o.steps[i].key)
	}
	return fmt.Sprintf("%s.steps[%d]", o.desc, i)
}

func (o *operator) expand(in interface{}) (interface{}, error) {
	store := o.store.toMap()
	store["string"] = func(in interface{}) string { return fmt.Sprintf("%v", in) }
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
				if numberRe.MatchString(v) {
					s = fmt.Sprintf("'%s'", v)
				} else {
					s = v
				}
			case int64:
				s = strconv.Itoa(int(v))
			case int:
				s = strconv.Itoa(v)
			default:
				reperr = fmt.Errorf("invalid format: %v\n%s", o, string(b))
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

type operators struct {
	ops []*operator
	t   *testing.T
}

func Load(pathp string, opts ...Option) (*operators, error) {
	ops := &operators{}
	books, err := Books(pathp)
	if err != nil {
		return nil, err
	}
	for _, b := range books {
		o, err := New(append(opts, b)...)
		if err != nil {
			return nil, err
		}
		if o.t != nil {
			ops.t = o.t
		}
		ops.ops = append(ops.ops, o)
	}
	return ops, nil
}

func (ops *operators) RunN(ctx context.Context) error {
	if ops.t != nil {
		ops.t.Helper()
	}
	for _, o := range ops.ops {
		if err := o.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
