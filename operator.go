package runn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/antonmedv/expr"
	"github.com/fatih/color"
	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/k1LoW/expand"
)

var (
	cyan     = color.New(color.FgCyan).SprintFunc()
	yellow   = color.New(color.FgYellow).SprintFunc()
	expandRe = regexp.MustCompile(`"?{{\s*([^}]+)\s*}}"?`)
	numberRe = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?$`)
)

type step struct {
	key           string
	runnerKey     string
	desc          string
	cond          string
	retry         *Retry
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
	bindRunner    *bindRunner
	bindCond      map[string]string
	includeRunner *includeRunner
	includePath   string
	debug         bool
}

const (
	storeVarsKey  = "vars"
	storeStepsKey = "steps"
)

type store struct {
	steps    []map[string]interface{}
	stepMaps map[string]interface{}
	vars     map[string]interface{}
	funcs    map[string]interface{}
	bindVars map[string]interface{}
	useMaps  bool
}

func (s *store) toMap() map[string]interface{} {
	store := map[string]interface{}{}
	for k, v := range s.funcs {
		store[k] = v
	}
	store[storeVarsKey] = s.vars
	if s.useMaps {
		store[storeStepsKey] = s.stepMaps
	} else {
		store[storeStepsKey] = s.steps
	}
	for k, v := range s.bindVars {
		store[k] = v
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
	failFast    bool
	included    bool
	cond        string
	skipped     bool
	out         io.Writer
	bookPath    string
}

func (o *operator) record(v map[string]interface{}) {
	if o.useMaps && len(o.steps) > 0 {
		o.store.stepMaps[o.steps[len(o.store.stepMaps)].key] = v
		return
	}
	o.store.steps = append(o.store.steps, v)
}

func (o *operator) deleteLatestRecord() {
	if o.useMaps && len(o.steps) > 0 {
		delete(o.store.stepMaps, o.steps[len(o.store.stepMaps)-1].key)
		return
	}
	o.store.steps = o.store.steps[:len(o.store.steps)-1]
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
			funcs:    bk.Funcs,
			bindVars: map[string]interface{}{},
			useMaps:  useMaps,
		},
		useMaps:  useMaps,
		desc:     bk.Desc,
		debug:    bk.Debug,
		interval: bk.interval,
		t:        bk.t,
		failFast: bk.failFast,
		included: bk.included,
		cond:     bk.If,
		out:      os.Stderr,
		bookPath: bk.path,
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
		if k == includeRunnerKey || k == testRunnerKey || k == dumpRunnerKey || k == execRunnerKey || k == bindRunnerKey {
			return nil, fmt.Errorf("runner name '%s' is reserved for built-in runner", k)
		}
		if k == ifSectionKey || k == descSectionKey || k == retrySectionKey {
			return nil, fmt.Errorf("runner name '%s' is reserved for built-in section", k)
		}
		delete(bk.runnerErrs, k)

		switch vv := v.(type) {
		case string:
			switch {
			case strings.Index(vv, "https://") == 0 || strings.Index(vv, "http://") == 0:
				hc, err := newHTTPRunner(k, vv, o)
				if err != nil {
					bk.runnerErrs[k] = err
					continue
				}
				o.httpRunners[k] = hc
			default:
				dc, err := newDBRunner(k, vv, o)
				if err != nil {
					bk.runnerErrs[k] = err
					continue
				}
				o.dbRunners[k] = dc
			}
		case map[string]interface{}:
			tmp, err := yaml.Marshal(vv)
			if err != nil {
				bk.runnerErrs[k] = err
				continue
			}
			c := &RunnerConfig{}
			if err := yaml.Unmarshal(tmp, c); err != nil {
				bk.runnerErrs[k] = err
				continue
			}

			if c.OpenApi3DocLocation != "" && !strings.HasPrefix(c.OpenApi3DocLocation, "https://") && !strings.HasPrefix(c.OpenApi3DocLocation, "http://") && !strings.HasPrefix(c.OpenApi3DocLocation, "/") {
				c.OpenApi3DocLocation = filepath.Join(o.root, c.OpenApi3DocLocation)
			}

			if c.Endpoint != "" {
				// httpRunner
				hc, err := newHTTPRunner(k, c.Endpoint, o)
				if err != nil {
					bk.runnerErrs[k] = err
					continue
				}
				hv, err := NewHttpValidator(c)
				if err != nil {
					bk.runnerErrs[k] = err
					continue
				}
				hc.validator = hv
				o.httpRunners[k] = hc
			}
		}
	}
	for k, v := range bk.httpRunners {
		delete(bk.runnerErrs, k)
		v.operator = o
		o.httpRunners[k] = v
	}
	for k, v := range bk.dbRunners {
		delete(bk.runnerErrs, k)
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

	var merr error
	for k, err := range bk.runnerErrs {
		merr = multierror.Append(merr, fmt.Errorf("runner %s error: %w", k, err))
	}
	if merr != nil {
		return nil, merr
	}

	for i, s := range bk.Steps {
		if err := validateStepKeys(s); err != nil {
			return nil, fmt.Errorf("invalid steps[%d]. %w: %s", i, err, s)
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

func validateStepKeys(s map[string]interface{}) error {
	if len(s) == 0 {
		return errors.New("step must specify at least one runner")
	}
	custom := 0
	for k := range s {
		if k == testRunnerKey || k == dumpRunnerKey || k == bindRunnerKey || k == ifSectionKey || k == descSectionKey || k == retrySectionKey {
			continue
		}
		custom += 1
	}
	if custom > 1 {
		return errors.New("runners that cannot be running at the same time are specified")
	}
	return nil
}

func (o *operator) AppendStep(key string, s map[string]interface{}) error {
	if o.t != nil {
		o.t.Helper()
	}
	step := &step{key: key, debug: o.debug}
	// if section
	if v, ok := s[ifSectionKey]; ok {
		step.cond, ok = v.(string)
		if !ok {
			return fmt.Errorf("invalid if condition: %v", v)
		}
		delete(s, ifSectionKey)
	}
	// desc section
	if v, ok := s[descSectionKey]; ok {
		step.desc, ok = v.(string)
		if !ok {
			return fmt.Errorf("invalid desc: %v", v)
		}
		delete(s, descSectionKey)
	}
	// retry section
	if v, ok := s[retrySectionKey]; ok {
		r, err := newRetry(v)
		if err != nil {
			return fmt.Errorf("invalid retry: %w\n%v", err, v)
		}
		step.retry = r
		delete(s, retrySectionKey)
	}
	// test runner
	if v, ok := s[testRunnerKey]; ok {
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
		delete(s, testRunnerKey)
	}
	// dump runner
	if v, ok := s[dumpRunnerKey]; ok {
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
		delete(s, dumpRunnerKey)
	}
	// bind runner
	if v, ok := s[bindRunnerKey]; ok {
		br, err := newBindRunner(o)
		if err != nil {
			return err
		}
		step.bindRunner = br
		vv, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid bind condition: %v", v)
		}
		cond := map[string]string{}
		for k, vvv := range vv {
			s, ok := vvv.(string)
			if !ok {
				return fmt.Errorf("invalid bind condition: %v", v)
			}
			cond[k] = s
		}
		step.bindCond = cond
		delete(s, bindRunnerKey)
	}

	k, v, ok := pop(s)
	if ok {
		step.runnerKey = k
		switch {
		case k == includeRunnerKey:
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
		case k == execRunnerKey:
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
		default:
			h, ok := o.httpRunners[k]
			if ok {
				step.httpRunner = h
				vv, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid http request: %v", v)
				}
				step.httpRequest = vv
			} else {
				db, ok := o.dbRunners[k]
				if ok {
					step.dbRunner = db
					vv, ok := v.(map[string]interface{})
					if !ok {
						return fmt.Errorf("invalid db query: %v", v)
					}
					step.dbQuery = vv
				} else {
					return fmt.Errorf("can not find client: %s", k)
				}
			}
		}
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
	// if
	if o.cond != "" {
		store := o.store.toMap()
		store["included"] = o.included
		tf, err := expr.Eval(fmt.Sprintf("(%s) == true", o.cond), store)
		if err != nil {
			return err
		}
		if !tf.(bool) {
			o.Debugf(yellow("Skip %s\n"), o.desc)
			o.skipped = true
			return nil
		}
	}

	for i, s := range o.steps {
		if i != 0 {
			time.Sleep(o.interval)
			o.Debugln("")
		}
		if s.cond != "" {
			store := o.store.toMap()
			store["included"] = o.included
			tf, err := expr.Eval(fmt.Sprintf("(%s) == true", s.cond), store)
			if err != nil {
				return err
			}
			if !tf.(bool) {
				if s.desc != "" {
					o.Debugf(yellow("Skip '%s' on %s\n"), s.desc, o.stepName(i))
				} else if s.runnerKey != "" {
					o.Debugf(yellow("Skip '%s' on %s\n"), s.runnerKey, o.stepName(i))
				} else {
					o.Debugf(yellow("Skip on %s\n"), o.stepName(i))
				}
				continue
			}
		}
		if s.runnerKey != "" {
			o.Debugf(cyan("Run '%s' on %s\n"), s.runnerKey, o.stepName(i))
		}

		stepFn := func() error {
			runned := false
			switch {
			case s.httpRunner != nil && s.httpRequest != nil:
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
				runned = true
			case s.dbRunner != nil && s.dbQuery != nil:
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
				runned = true
			case s.execRunner != nil && s.execCommand != nil:
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
				runned = true
			case s.includeRunner != nil && s.includePath != "":
				if err := s.includeRunner.Run(ctx, s.includePath); err != nil {
					return fmt.Errorf("include failed on %s: %v", o.stepName(i), err)
				}
				runned = true
			}
			// dump runner
			if s.dumpRunner != nil && s.dumpCond != "" {
				o.Debugf(cyan("Run '%s' on %s\n"), dumpRunnerKey, o.stepName(i))
				if err := s.dumpRunner.Run(ctx, s.dumpCond); err != nil {
					return fmt.Errorf("dump failed on %s: %v", o.stepName(i), err)
				}
				if !runned {
					o.record(nil)
					runned = true
				}
			}
			// bind runner
			if s.bindRunner != nil && s.bindCond != nil {
				o.Debugf(cyan("Run '%s' on %s\n"), bindRunnerKey, o.stepName(i))
				if err := s.bindRunner.Run(ctx, s.bindCond); err != nil {
					return fmt.Errorf("bind failed on %s: %v", o.stepName(i), err)
				}
				if !runned {
					o.record(nil)
					runned = true
				}
			}
			// test runner
			if s.testRunner != nil && s.testCond != "" {
				o.Debugf(cyan("Run '%s' on %s\n"), testRunnerKey, o.stepName(i))
				if err := s.testRunner.Run(ctx, s.testCond); err != nil {
					return fmt.Errorf("test failed on %s: %v", o.stepName(i), err)
				}
				if !runned {
					o.record(nil)
					runned = true
				}
			}

			if !runned {
				return fmt.Errorf("invalid runner: %v", o.stepName(i))
			}
			return nil
		}

		// retry
		if s.retry != nil {
			success := false
			var t string
			for s.retry.Retry(ctx) {
				if err := stepFn(); err != nil {
					return err
				}
				store := o.store.toMap()
				t = buildTree(s.retry.Until, store)
				o.Debugln("-----START RETRY CONDITION-----")
				o.Debugf("%s", t)
				o.Debugln("-----END RETRY CONDITION-----")
				tf, err := expr.Eval(fmt.Sprintf("(%s) == true", s.retry.Until), store)
				if err != nil {
					return err
				}
				if tf.(bool) {
					success = true
					break
				}
				o.deleteLatestRecord()
			}
			if !success {
				err := fmt.Errorf("(%s) is not true\n%s", s.retry.Until, t)
				return fmt.Errorf("retry failed on %s: %w", o.stepName(i), err)
			}
		} else {
			if err := stepFn(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *operator) stepName(i int) string {
	if o.useMaps {
		return fmt.Sprintf("'%s'.steps.%s", o.desc, o.steps[i].key)
	}
	return fmt.Sprintf("'%s'.steps[%d]", o.desc, i)
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

func (o *operator) Debugln(a string) {
	if !o.debug {
		return
	}
	_, _ = fmt.Fprintln(o.out, a)
}

func (o *operator) Debugf(format string, a ...interface{}) {
	if !o.debug {
		return
	}
	_, _ = fmt.Fprintf(o.out, format, a...)
}

func (o *operator) Skipped() bool {
	return o.skipped
}

type operators struct {
	ops []*operator
	t   *testing.T
}

func Load(pathp string, opts ...Option) (*operators, error) {
	bk := newBook()
	for _, opt := range opts {
		if err := opt(bk); err != nil {
			return nil, err
		}
	}
	ops := &operators{}
	books, err := Books(pathp)
	if err != nil {
		return nil, err
	}
	skipPaths := []string{}
	om := map[string]*operator{}
	for _, b := range books {
		o, err := New(append(opts, b)...)
		if err != nil {
			return nil, err
		}
		if bk.skipIncluded {
			for _, s := range o.steps {
				if s.includeRunner != nil {
					skipPaths = append(skipPaths, filepath.Join(o.root, s.includePath))
				}
			}
		}
		if o.t != nil {
			ops.t = o.t
		}
		om[o.bookPath] = o
	}

	for p, o := range om {
		if contains(skipPaths, p) {
			o.Debugf(yellow("Skip %s because it is already included from another runbook\n"), p)
			continue
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
		if err := o.Run(ctx); err != nil && o.failFast {
			return err
		}
	}
	return nil
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func pop(s map[string]interface{}) (string, interface{}, bool) {
	for k, v := range s {
		delete(s, k)
		return k, v, true
	}
	return "", nil, false
}
