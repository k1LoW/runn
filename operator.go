package runn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/antonmedv/expr"
	"github.com/fatih/color"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
	"github.com/k1LoW/stopw"
	"github.com/rs/xid"
	"go.uber.org/multierr"
)

const (
	delimStart = "{{"
	delimEnd   = "}}"
)

var (
	cyan     = color.New(color.FgCyan).SprintFunc()
	yellow   = color.New(color.FgYellow).SprintFunc()
	expandRe = regexp.MustCompile(fmt.Sprintf(`"?%s\s*([^}]+)\s*%s"?`, delimStart, delimEnd))
	numberRe = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?$`)
)

type step struct {
	key           string
	runnerKey     string
	desc          string
	cond          string
	loop          *Loop
	httpRunner    *httpRunner
	httpRequest   map[string]interface{}
	dbRunner      *dbRunner
	dbQuery       map[string]interface{}
	grpcRunner    *grpcRunner
	grpcRequest   map[string]interface{}
	execRunner    *execRunner
	execCommand   map[string]interface{}
	testRunner    *testRunner
	testCond      string
	dumpRunner    *dumpRunner
	dumpCond      string
	bindRunner    *bindRunner
	bindCond      map[string]string
	includeRunner *includeRunner
	includeConfig *includeConfig
	parent        *operator
	debug         bool
}

func (s *step) ids() []string {
	var ids []string
	if s.parent != nil {
		ids = s.parent.ids()
	}
	ids = append(ids, s.key)
	return ids
}

type operator struct {
	id          string
	httpRunners map[string]*httpRunner
	dbRunners   map[string]*dbRunner
	grpcRunners map[string]*grpcRunner
	steps       []*step
	store       store
	desc        string
	useMap      bool // Use map syntax in `steps:`.
	debug       bool
	profile     bool
	interval    time.Duration
	root        string
	t           *testing.T
	thisT       *testing.T
	parent      *step
	failFast    bool
	included    bool
	cond        string
	skipTest    bool
	skipped     bool
	out         io.Writer
	bookPath    string
	beforeFuncs []func() error
	afterFuncs  []func() error
	sw          *stopw.Span
}

func (o *operator) record(v map[string]interface{}) {
	if o.useMap {
		o.recordToMap(v)
		return
	}
	o.recordToArray(v)
}

func (o *operator) recordToArray(v map[string]interface{}) {
	if o.store.loopIndex != nil && *o.store.loopIndex > 0 {
		// delete values of prevous loop
		o.store.steps = o.store.steps[:len(o.store.steps)-1]
	}
	o.store.recordToArray(v)
}

func (o *operator) recordToMap(v map[string]interface{}) {
	if o.store.loopIndex != nil && *o.store.loopIndex > 0 {
		// delete values of prevous loop
		delete(o.store.stepMaps, o.steps[len(o.store.stepMaps)-1].key)
	}
	k := o.steps[len(o.store.stepMaps)].key
	o.store.recordToMap(k, v)
}

func (o *operator) Close() {
	for _, r := range o.grpcRunners {
		_ = r.Close()
	}
}

func (o *operator) ids() []string {
	var ids []string
	if o.parent != nil {
		ids = o.parent.ids()
	}
	ids = append(ids, o.id)
	return ids
}

func New(opts ...Option) (*operator, error) {
	bk := newBook()
	if err := bk.applyOptions(opts...); err != nil {
		return nil, err
	}

	useMap := false
	if len(bk.stepKeys) > 0 && len(bk.stepKeys) == len(bk.Steps) {
		useMap = true
	}

	o := &operator{
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
		grpcRunners: map[string]*grpcRunner{},
		store: store{
			steps:    []map[string]interface{}{},
			stepMaps: map[string]interface{}{},
			vars:     bk.Vars,
			funcs:    bk.Funcs,
			bindVars: map[string]interface{}{},
			useMap:   useMap,
		},
		useMap:      useMap,
		desc:        bk.Desc,
		debug:       bk.Debug,
		profile:     bk.profile,
		interval:    bk.interval,
		t:           bk.t,
		thisT:       bk.t,
		failFast:    bk.failFast,
		included:    bk.included,
		cond:        bk.If,
		skipTest:    bk.SkipTest,
		out:         os.Stderr,
		bookPath:    bk.path,
		beforeFuncs: bk.beforeFuncs,
		afterFuncs:  bk.afterFuncs,
		sw:          stopw.New(),
	}

	if bk.path != "" {
		o.id = bk.path
		o.root = filepath.Dir(bk.path)
	} else {
		o.id = xid.New().String()
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		o.root = wd
	}

	for k, v := range bk.Runners {
		if k == deprecatedRetrySectionKey {
			o.Warnf("'%s' is deprecated. use %s instead", deprecatedRetrySectionKey, loopSectionKey)
		}
		if k == includeRunnerKey || k == testRunnerKey || k == dumpRunnerKey || k == execRunnerKey || k == bindRunnerKey {
			return nil, fmt.Errorf("runner name '%s' is reserved for built-in runner", k)
		}
		if k == ifSectionKey || k == descSectionKey || k == loopSectionKey || k == deprecatedRetrySectionKey {
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
			case strings.Index(vv, "grpc://") == 0:
				addr := strings.TrimPrefix(vv, "grpc://")
				gc, err := newGrpcRunner(k, addr, o)
				if err != nil {
					bk.runnerErrs[k] = err
					continue
				}
				o.grpcRunners[k] = gc
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
			detect := false

			// HTTP Runner
			c := &httpRunnerConfig{}
			if err := yaml.Unmarshal(tmp, c); err == nil {
				if c.Endpoint != "" {
					detect = true
					r, err := newHTTPRunner(k, c.Endpoint, o)
					if err != nil {
						bk.runnerErrs[k] = err
						continue
					}
					if c.OpenApi3DocLocation != "" && !strings.HasPrefix(c.OpenApi3DocLocation, "https://") && !strings.HasPrefix(c.OpenApi3DocLocation, "http://") && !strings.HasPrefix(c.OpenApi3DocLocation, "/") {
						c.OpenApi3DocLocation = filepath.Join(o.root, c.OpenApi3DocLocation)
					}
					hv, err := newHttpValidator(c)
					if err != nil {
						bk.runnerErrs[k] = err
						continue
					}
					r.validator = hv
					o.httpRunners[k] = r
				}
			}

			// gRPC Runner
			if !detect {
				c := &grpcRunnerConfig{}
				if err := yaml.Unmarshal(tmp, c); err == nil {
					if c.Addr != "" {
						detect = true
						r, err := newGrpcRunner(k, c.Addr, o)
						if err != nil {
							bk.runnerErrs[k] = err
							continue
						}
						r.tls = c.TLS
						if c.cacert != nil {
							r.cacert = c.cacert
						} else if strings.HasPrefix(c.CACert, "/") {
							b, err := os.ReadFile(c.CACert)
							if err != nil {
								return nil, err
							}
							r.cacert = b
						} else {
							b, err := os.ReadFile(filepath.Join(o.root, c.CACert))
							if err != nil {
								return nil, err
							}
							r.cacert = b
						}
						if c.cert != nil {
							r.cert = c.cert
						} else if strings.HasPrefix(c.Cert, "/") {
							b, err := os.ReadFile(c.Cert)
							if err != nil {
								return nil, err
							}
							r.cert = b
						} else {
							b, err := os.ReadFile(filepath.Join(o.root, c.Cert))
							if err != nil {
								return nil, err
							}
							r.cert = b
						}
						if c.key != nil {
							r.key = c.key
						} else if strings.HasPrefix(c.Key, "/") {
							b, err := os.ReadFile(c.Key)
							if err != nil {
								return nil, err
							}
							r.key = b
						} else {
							b, err := os.ReadFile(filepath.Join(o.root, c.Key))
							if err != nil {
								return nil, err
							}
							r.key = b
						}
						r.skipVerify = c.SkipVerify
						o.grpcRunners[k] = r
					}
				}
			}

			if !detect {
				bk.runnerErrs[k] = fmt.Errorf("cannot detect runner: %s", string(tmp))
				continue
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
	for k, v := range bk.grpcRunners {
		delete(bk.runnerErrs, k)
		v.operator = o
		o.grpcRunners[k] = v
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
	for k := range o.grpcRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names: %s", k)
		}
		keys[k] = struct{}{}
	}

	var merr error
	for k, err := range bk.runnerErrs {
		merr = multierr.Append(merr, fmt.Errorf("runner %s error: %w", k, err))
	}
	if merr != nil {
		return nil, merr
	}

	for i, s := range bk.Steps {
		if err := validateStepKeys(s); err != nil {
			return nil, fmt.Errorf("invalid steps[%d]. %w: %s", i, err, s)
		}
		key := fmt.Sprintf("%d", i)
		if o.useMap {
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
		if k == testRunnerKey || k == dumpRunnerKey || k == bindRunnerKey || k == ifSectionKey || k == descSectionKey || k == loopSectionKey || k == deprecatedRetrySectionKey {
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
	step := &step{key: key, parent: o, debug: o.debug}
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
	// loop section
	if v, ok := s[loopSectionKey]; ok {
		r, err := newLoop(v)
		if err != nil {
			return fmt.Errorf("invalid loop: %w\n%v", err, v)
		}
		step.loop = r
		delete(s, loopSectionKey)
	}
	// deprecated `retry:`
	if v, ok := s[deprecatedRetrySectionKey]; ok {
		r, err := newLoop(v)
		if err != nil {
			return fmt.Errorf("invalid loop: %w\n%v", err, v)
		}
		step.loop = r
		delete(s, deprecatedRetrySectionKey)
	}
	// test runner
	if v, ok := s[testRunnerKey]; ok {
		tr, err := newTestRunner(o)
		if err != nil {
			return err
		}
		step.testRunner = tr
		switch vv := v.(type) {
		case bool:
			if vv {
				step.testCond = "true"
			} else {
				step.testCond = "false"
			}
		case string:
			step.testCond = vv
		default:
			return fmt.Errorf("invalid test condition: %v", v)
		}
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
			c, err := parseIncludeConfig(v)
			if err != nil {
				return err
			}
			c.step = step
			step.includeConfig = c
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
			detected := false
			h, ok := o.httpRunners[k]
			if ok {
				step.httpRunner = h
				vv, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid http request: %v", v)
				}
				step.httpRequest = vv
				detected = true
			}
			db, ok := o.dbRunners[k]
			if ok && !detected {
				step.dbRunner = db
				vv, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid db query: %v", v)
				}
				step.dbQuery = vv
				detected = true
			}
			gc, ok := o.grpcRunners[k]
			if ok && !detected {
				step.grpcRunner = gc
				vv, ok := v.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid gRPC request: %v", v)
				}
				step.grpcRequest = vv
				detected = true
			}
			if !detected {
				return fmt.Errorf("cannot find client: %s", k)
			}
		}
	}
	o.steps = append(o.steps, step)
	return nil
}

func (o *operator) Run(ctx context.Context) error {
	if o.t != nil {
		o.t.Helper()
	}
	if !o.profile {
		o.sw.Disable()
	}
	defer o.sw.Start().Stop()
	defer o.Close()
	return o.run(ctx)
}

func (o *operator) DumpProfile(w io.Writer) error {
	r := o.sw.Result()
	if r == nil {
		return errors.New("no profile")
	}
	enc := json.NewEncoder(w)
	if err := enc.Encode(r); err != nil {
		return err
	}
	return nil
}

func (o *operator) run(ctx context.Context) error {
	defer o.sw.Start(toInterfaces(o.ids())...).Stop()
	if o.t != nil {
		o.t.Helper()
		var err error
		o.t.Run(o.testName(), func(t *testing.T) {
			t.Helper()
			o.thisT = t
			err = o.runInternal(ctx)
			if err != nil {
				t.Error(err)
			}
		})
		o.thisT = o.t
		if err != nil {
			return fmt.Errorf("failed to run %s: %w", o.id, err)
		}
		return nil
	}
	if err := o.runInternal(ctx); err != nil {
		return fmt.Errorf("failed to run %s: %w", o.id, err)
	}
	return nil
}

func (o *operator) runInternal(ctx context.Context) error {
	if o.t != nil {
		o.t.Helper()
	}
	// if
	if o.cond != "" {
		store := o.store.toMap()
		store[storeIncludedKey] = o.included
		tf, err := evalCond(o.cond, store)
		if err != nil {
			return err
		}
		if !tf {
			o.Debugf(yellow("Skip %s\n"), o.desc)
			o.skipped = true
			return nil
		}
	}
	// beforeFuncs
	for i, fn := range o.beforeFuncs {
		ids := append(o.ids(), fmt.Sprintf("beforeFuncs[%d]", i))
		o.sw.Start(toInterfaces(ids)...)
		if err := fn(); err != nil {
			o.sw.Stop(toInterfaces(ids)...)
			return err
		}
		o.sw.Stop(toInterfaces(ids)...)
	}

	// steps
	for i, s := range o.steps {
		err := func() error {
			ids := append(o.ids(), s.key)
			defer o.sw.Start(toInterfaces(ids)...).Stop()
			if i != 0 {
				time.Sleep(o.interval)
				o.Debugln("")
			}
			if s.cond != "" {
				store := o.store.toMap()
				store[storeIncludedKey] = o.included
				tf, err := evalCond(s.cond, store)
				if err != nil {
					return err
				}
				if !tf {
					if s.desc != "" {
						o.Debugf(yellow("Skip '%s' on %s\n"), s.desc, o.stepName(i))
					} else if s.runnerKey != "" {
						o.Debugf(yellow("Skip '%s' on %s\n"), s.runnerKey, o.stepName(i))
					} else {
						o.Debugf(yellow("Skip on %s\n"), o.stepName(i))
					}
					o.record(nil)
					return nil
				}
			}
			if s.runnerKey != "" {
				o.Debugf(cyan("Run '%s' on %s\n"), s.runnerKey, o.stepName(i))
			}

			stepFn := func(t *testing.T) error {
				if t != nil {
					t.Helper()
				}
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
				case s.grpcRunner != nil && s.grpcRequest != nil:
					req, err := parseGrpcRequest(s.grpcRequest, o.expand)
					if err != nil {
						return fmt.Errorf("invalid %s: %v", o.stepName(i), s.grpcRequest)
					}
					if err := s.grpcRunner.Run(ctx, req); err != nil {
						return fmt.Errorf("gRPC request failed on %s: %v", o.stepName(i), err)
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
				case s.includeRunner != nil && s.includeConfig != nil:
					if err := s.includeRunner.Run(ctx, s.includeConfig); err != nil {
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
					if o.skipTest {
						o.Debugf(yellow("Skip '%s' on %s\n"), testRunnerKey, o.stepName(i))
						if !runned {
							o.record(nil)
						}
						return nil
					}
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

			// loop
			if s.loop != nil {
				retrySuccess := false
				if s.loop.Until == "" {
					retrySuccess = true
				}
				var t string
				var i int
				c, err := evalCount(s.loop.Count, o.store.toMap())
				if err != nil {
					return err
				}
				for s.loop.Loop(ctx) {
					if i >= c {
						break
					}
					ii := i
					o.store.loopIndex = &ii
					if err := stepFn(o.thisT); err != nil {
						o.store.loopIndex = nil
						return err
					}
					if s.loop.Until != "" {
						store := o.store.toMap()
						t = buildTree(s.loop.Until, store)
						o.Debugln("-----START LOOP CONDITION-----")
						o.Debugf("%s", t)
						o.Debugln("-----END LOOP CONDITION-----")
						tf, err := evalCond(s.loop.Until, store)
						if err != nil {
							o.store.loopIndex = nil
							return err
						}
						if tf {
							retrySuccess = true
							break
						}
					}
					i++
				}
				o.store.loopIndex = nil
				if !retrySuccess {
					err := fmt.Errorf("(%s) is not true\n%s", s.loop.Until, t)
					return fmt.Errorf("loop failed on %s: %w", o.stepName(i), err)
				}
			} else {
				if err := stepFn(o.thisT); err != nil {
					return err
				}
			}
			return nil
		}()

		if err != nil {
			return err
		}
	}

	// afterFuncs
	for i, fn := range o.afterFuncs {
		ids := append(o.ids(), fmt.Sprintf("afterFuncs[%d]", i))
		o.sw.Start(toInterfaces(ids)...)
		if err := fn(); err != nil {
			o.sw.Stop(toInterfaces(ids)...)
			return err
		}
		o.sw.Stop(toInterfaces(ids)...)
	}

	return nil
}

func (o *operator) testName() string {
	if o.bookPath == "" {
		return fmt.Sprintf("%s(-)", o.desc)
	}
	return fmt.Sprintf("%s(%s)", o.desc, o.bookPath)
}

func (o *operator) stepName(i int) string {
	if o.useMap {
		return fmt.Sprintf("'%s'.steps.%s", o.desc, o.steps[i].key)
	}
	return fmt.Sprintf("'%s'.steps[%d]", o.desc, i)
}

func (o *operator) expand(in interface{}) (interface{}, error) {
	store := o.store.toMap()
	b, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	var reperr error
	replacefunc := func(in string) string {
		if !strings.Contains(in, delimStart) {
			return in
		}
		matches := expandRe.FindAllStringSubmatch(in, -1)
		oldnew := []string{}
		for _, m := range matches {
			o, err := expr.Eval(m[1], store)
			if err != nil {
				reperr = err
				return ""
			}
			var s string
			switch v := o.(type) {
			case string:
				// Stringify only one expression.
				if strings.TrimSpace(in) == m[0] && numberRe.MatchString(v) {
					s = fmt.Sprintf("'%s'", v)
				} else {
					s = v
				}
			case int64:
				s = strconv.Itoa(int(v))
			case float64:
				s = strconv.FormatFloat(v, 'f', -1, 64)
			case int:
				s = strconv.Itoa(v)
			case bool:
				s = strconv.FormatBool(v)
			case map[string]interface{}, []interface{}:
				bytes, err := json.Marshal(v)
				if err != nil {
					reperr = fmt.Errorf("json.Marshal error: %w", err)
				} else {
					s = string(bytes)
				}
			default:
				reperr = fmt.Errorf("invalid format: evaluated %s, but got %T(%v)", m[1], o, o)
				return ""
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

func (o *operator) Warnf(format string, a ...interface{}) {
	_, _ = fmt.Fprintf(o.out, format, a...)
}

func (o *operator) Skipped() bool {
	return o.skipped
}

type operators struct {
	ops     []*operator
	t       *testing.T
	sw      *stopw.Span
	profile bool
}

func Load(pathp string, opts ...Option) (*operators, error) {
	bk := newBook()
	opts = append([]Option{RunMatch(os.Getenv("RUNN_RUN"))}, opts...)
	if err := bk.applyOptions(opts...); err != nil {
		return nil, err
	}

	sw := stopw.New()
	ops := &operators{
		t:       bk.t,
		sw:      sw,
		profile: bk.profile,
	}
	books, err := Books(pathp)
	if err != nil {
		return nil, err
	}
	skipPaths := []string{}
	om := map[string]*operator{}
	for _, b := range books {
		o, err := New(append([]Option{b}, opts...)...)
		if err != nil {
			return nil, err
		}
		if bk.skipIncluded {
			for _, s := range o.steps {
				if s.includeRunner != nil && s.includeConfig != nil {
					skipPaths = append(skipPaths, filepath.Join(o.root, s.includeConfig.path))
				}
			}
		}
		om[o.bookPath] = o
	}

	for p, o := range om {
		if !bk.runMatch.MatchString(p) {
			o.Debugf(yellow("Skip %s because it does not match %s\n"), p, bk.runMatch.String())
			continue
		}
		if contains(skipPaths, p) {
			o.Debugf(yellow("Skip %s because it is already included from another runbook\n"), p)
			continue
		}
		o.sw = ops.sw
		ops.ops = append(ops.ops, o)
	}
	if bk.runShardN > 0 {
		ops.ops = partOperators(ops.ops, bk.runShardN, bk.runShardIndex)
	}
	if bk.runSample > 0 {
		ops.ops = sampleOperators(ops.ops, bk.runSample)
	}
	return ops, nil
}

func (ops *operators) RunN(ctx context.Context) error {
	if ops.t != nil {
		ops.t.Helper()
	}
	if !ops.profile {
		ops.sw.Disable()
	}
	defer ops.sw.Start().Stop()
	defer ops.Close()
	for _, o := range ops.ops {
		if err := o.run(ctx); err != nil && o.failFast {
			return err
		}
	}
	return nil
}

func (ops *operators) Close() {
	for _, o := range ops.ops {
		o.Close()
	}
}

func (ops *operators) DumpProfile(w io.Writer) error {
	r := ops.sw.Result()
	if r == nil {
		return errors.New("no profile")
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return err
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

func partOperators(ops []*operator, n, i int) []*operator {
	all := make([]*operator, len(ops))
	copy(all, ops)
	sortOperators(all)
	var part []*operator
	for ii, o := range all {
		if math.Mod(float64(ii), float64(n)) == float64(i) {
			part = append(part, o)
		}
	}
	return part
}

func sortOperators(ops []*operator) {
	sort.SliceStable(ops, func(i, j int) bool {
		if ops[i].bookPath == ops[j].bookPath {
			return ops[i].desc < ops[j].desc
		}
		return ops[i].bookPath < ops[j].bookPath
	})
}

func sampleOperators(ops []*operator, num int) []*operator {
	if len(ops) <= num {
		return ops
	}
	rand.Seed(time.Now().UnixNano())
	var sample []*operator
	n := make([]*operator, len(ops))
	copy(n, ops)

	for i := 0; i < num; i++ {
		idx := rand.Intn(len(n))
		sample = append(sample, n[idx])
		n = append(n[:idx], n[idx+1:]...)
	}
	return sample
}

func pop(s map[string]interface{}) (string, interface{}, bool) {
	for k, v := range s {
		defer delete(s, k)
		return k, v, true
	}
	return "", nil, false
}

func toInterfaces(in []string) []interface{} {
	s := make([]interface{}, len(in))
	for i, v := range in {
		s[i] = v
	}
	return s
}
