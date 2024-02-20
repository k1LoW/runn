package runn

import (
	"context"
	ejson "encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/k1LoW/concgroup"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/stopw"
	"github.com/ryo-yamaoka/otchkiss"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"go.uber.org/multierr"
)

var errStepSkiped = errors.New("step skipped")

var _ otchkiss.Requester = (*operators)(nil)

type operator struct {
	id             string
	httpRunners    map[string]*httpRunner
	dbRunners      map[string]*dbRunner
	grpcRunners    map[string]*grpcRunner
	cdpRunners     map[string]*cdpRunner
	sshRunners     map[string]*sshRunner
	includeRunners map[string]*includeRunner
	steps          []*step
	store          store
	desc           string
	labels         []string
	useMap         bool // Use map syntax in `steps:`.
	debug          bool
	profile        bool
	interval       time.Duration
	loop           *Loop
	// loopIndex - Index of the loop is dynamically recorded at runtime
	loopIndex   *int
	concurrency []string
	// root - Root directory of runbook ( rubbook path or working directory )
	root        string
	t           *testing.T
	thisT       *testing.T
	parent      *step
	force       bool
	trace       bool
	waitTimeout time.Duration
	failFast    bool
	included    bool
	ifCond      string
	skipTest    bool
	skipped     bool
	stdout      io.Writer
	stderr      io.Writer
	// Skip some errors for `runn list`
	newOnly  bool
	bookPath string
	// Number of steps for `runn list`
	numberOfSteps int
	beforeFuncs   []func(*RunResult) error
	afterFuncs    []func(*RunResult) error
	sw            *stopw.Span
	capturers     capturers
	runResult     *RunResult

	mu sync.Mutex
}

// ID returns id of runbook.
func (o *operator) ID() string {
	return o.id
}

// runbookID returns id of the root runbook.
func (o *operator) runbookID() string { //nolint:unused
	return o.trails().runbookID()
}

// Desc returns `desc:` of runbook.
func (o *operator) Desc() string {
	return o.desc
}

// If returns `if:` of runbook.
func (o *operator) If() string {
	return o.ifCond
}

// BookPath returns path of runbook.
func (o *operator) BookPath() string {
	return o.bookPath
}

// NumberOfSteps returns number of steps.
func (o *operator) NumberOfSteps() int {
	return o.numberOfSteps
}

// Close runners.
func (o *operator) Close(force bool) {
	for _, r := range o.grpcRunners {
		if !force && r.target == "" {
			continue
		}
		_ = r.Close()
	}
	for _, r := range o.cdpRunners {
		_ = r.Close()
	}
	for _, r := range o.sshRunners {
		_ = r.Close()
	}
	for _, r := range o.dbRunners {
		if !force && r.dsn == "" {
			continue
		}
		_ = r.Close()
	}
}

func (o *operator) runStep(ctx context.Context, i int, s *step) error {
	if o.t != nil {
		o.t.Helper()
	}
	trs := s.trails()
	o.capturers.setCurrentTrails(trs)
	defer o.sw.Start(trs.toProfileIDs()...).Stop()
	if i != 0 {
		// interval:
		time.Sleep(o.interval)
		o.Debugln("")
	}
	if s.ifCond != "" {
		tf, err := o.expandCondBeforeRecord(s.ifCond)
		if err != nil {
			return err
		}
		if !tf {
			if s.desc != "" {
				o.Debugf(yellow("Skip %q on %s\n"), s.desc, o.stepName(i))
			} else if s.runnerKey != "" {
				o.Debugf(yellow("Skip %q on %s\n"), s.runnerKey, o.stepName(i))
			} else {
				o.Debugf(yellow("Skip on %s\n"), o.stepName(i))
			}
			return errStepSkiped
		}
	}
	if s.desc != "" {
		o.Debugf(cyan("Run %q on %s\n"), s.desc, o.stepName(i))
	} else if s.runnerKey != "" {
		o.Debugf(cyan("Run %q on %s\n"), s.runnerKey, o.stepName(i))
	}

	stepFn := func(t *testing.T) error {
		s.clearResult()
		if t != nil {
			t.Helper()
		}
		run := false
		switch {
		case s.httpRunner != nil && s.httpRequest != nil:
			if err := s.httpRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("http request failed on %s: %w", o.stepName(i), err)
			}
			run = true
		case s.dbRunner != nil && s.dbQuery != nil:
			if err := s.dbRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("db query failed on %s: %w", o.stepName(i), err)
			}
			run = true
		case s.grpcRunner != nil && s.grpcRequest != nil:
			if err := s.grpcRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("gRPC request failed on %s: %w", o.stepName(i), err)
			}
			run = true
		case s.cdpRunner != nil && s.cdpActions != nil:
			if err := s.cdpRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("cdp action failed on %s: %w", o.stepName(i), err)
			}
			run = true
		case s.sshRunner != nil && s.sshCommand != nil:
			if err := s.sshRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("ssh command failed on %s: %w", o.stepName(i), err)
			}
			run = true
		case s.execRunner != nil && s.execCommand != nil:
			if err := s.execRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("exec command failed on %s: %w", o.stepName(i), err)
			}
			run = true
		case s.includeRunner != nil && s.includeConfig != nil:
			if err := s.includeRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("include failed on %s: %w", o.stepName(i), err)
			}
			run = true
		}
		// dump runner
		if s.dumpRunner != nil && s.dumpRequest != nil {
			o.Debugf(cyan("Run %q on %s\n"), dumpRunnerKey, o.stepName(i))
			if err := s.dumpRunner.Run(ctx, s, !run); err != nil {
				return fmt.Errorf("dump failed on %s: %w", o.stepName(i), err)
			}
			run = true
		}
		// bind runner
		if s.bindRunner != nil && s.bindCond != nil {
			o.Debugf(cyan("Run %q on %s\n"), bindRunnerKey, o.stepName(i))
			if err := s.bindRunner.Run(ctx, s, !run); err != nil {
				return fmt.Errorf("bind failed on %s: %w", o.stepName(i), err)
			}
			run = true
		}
		// test runner
		if s.testRunner != nil && s.testCond != "" {
			if o.skipTest {
				o.Debugf(yellow("Skip %q on %s\n"), testRunnerKey, o.stepName(i))
				if !run {
					return errStepSkiped
				}
				return nil
			}
			o.Debugf(cyan("Run %q on %s\n"), testRunnerKey, o.stepName(i))
			if err := s.testRunner.Run(ctx, s, !run); err != nil {
				if s.desc != "" {
					return fmt.Errorf("test failed on %s %q: %w", o.stepName(i), s.desc, err)
				} else {
					return fmt.Errorf("test failed on %s: %w", o.stepName(i), err)
				}
			}
			run = true
		}

		if !run {
			return fmt.Errorf("invalid runner: %v", o.stepName(i))
		}
		return nil
	}

	// loop
	if s.loop != nil {
		defer func() {
			o.store.loopIndex = nil
			s.loopIndex = nil
		}()
		retrySuccess := false
		if s.loop.Until == "" {
			retrySuccess = true
		}
		var (
			bt string
			j  int
		)
		c, err := EvalCount(s.loop.Count, o.store.toMap())
		if err != nil {
			return err
		}
		for s.loop.Loop(ctx) {
			if j >= c {
				break
			}
			jj := j
			o.store.loopIndex = &jj
			s.loopIndex = &jj
			trs := s.trails()
			o.capturers.setCurrentTrails(trs)
			sw := o.sw.Start(trs.toProfileIDs()...)
			if err := stepFn(o.thisT); err != nil {
				sw.Stop()
				return fmt.Errorf("loop failed: %w", err)
			}
			sw.Stop()
			if s.loop.Until != "" {
				store := o.store.toMap()
				store[storeRootKeyIncluded] = o.included
				store[storeRootPrevious] = o.store.previous()
				store[storeRootKeyCurrent] = o.store.latest()
				bt, err = buildTree(s.loop.Until, store)
				if err != nil {
					return fmt.Errorf("loop failed on %s: %w", o.stepName(i), err)
				}
				tf, err := EvalCond(s.loop.Until, store)
				if err != nil {
					return fmt.Errorf("loop failed on %s: %w", o.stepName(i), err)
				}
				if tf {
					retrySuccess = true
					break
				}
			}
			j++
		}
		if !retrySuccess {
			err := fmt.Errorf("(%s) is not true\n%s", s.loop.Until, bt)
			o.store.loopIndex = nil
			if s.loop.interval != nil {
				return fmt.Errorf("retry loop failed on %s.loop (count: %d, interval: %v): %w", o.stepName(i), c, *s.loop.interval, err)
			} else {
				return fmt.Errorf("retry loop failed on %s.loop (count: %d, minInterval: %v, maxInterval: %v): %w", o.stepName(i), c, *s.loop.minInterval, *s.loop.maxInterval, err)
			}
		}
	} else {
		if err := stepFn(o.thisT); err != nil {
			return err
		}
	}
	return nil
}

// Record that it has not been run.
func (o *operator) recordNotRun(i int) {
	if o.store.length() == i+1 {
		return
	}
	v := map[string]any{}
	v[storeStepKeyRun] = false
	if o.useMap {
		o.recordAsMapped(v)
		return
	}
	o.recordAsListed(v)
}

func (o *operator) record(v map[string]any) {
	if v == nil {
		v = map[string]any{}
	}
	v[storeStepKeyRun] = true
	if o.useMap {
		o.recordAsMapped(v)
		return
	}
	o.recordAsListed(v)
}

func (o *operator) recordAsListed(v map[string]any) {
	if o.store.loopIndex != nil && *o.store.loopIndex > 0 {
		// delete values of prevous loop
		o.store.steps = o.store.steps[:o.store.length()-1]
	}
	o.store.recordAsListed(v)
}

func (o *operator) recordAsMapped(v map[string]any) {
	if o.store.loopIndex != nil && *o.store.loopIndex > 0 {
		// delete values of prevous loop
		o.store.removeLatestAsMapped()
	}
	// Get next key
	k := o.steps[o.store.length()].key
	o.store.recordAsMapped(k, v)
}

func (o *operator) recordToLatest(key string, value any) error {
	r := o.Result()
	r.StepResults = o.StepResults()
	o.capturers.captureResultByStep(o.trails(), r)
	return o.store.recordToLatest(key, value)
}

func (o *operator) recordToCookie(cookies []*http.Cookie) {
	o.store.recordToCookie(cookies)
}

func (o *operator) generateTrail() Trail {
	return Trail{
		Type:        TrailTypeRunbook,
		Desc:        o.desc,
		RunbookID:   o.id,
		RunbookPath: o.bookPath,
	}
}

func (o *operator) trails() Trails {
	var trs Trails
	if o.parent != nil {
		trs = o.parent.trails()
	}
	trs = append(trs, o.generateTrail())
	if o.loopIndex != nil {
		trs = append(trs, Trail{
			Type:      TrailTypeLoop,
			LoopIndex: o.loopIndex,
			RunbookID: o.id,
		})
	}
	return trs
}

// New returns *operator.
func New(opts ...Option) (*operator, error) {
	bk := newBook()
	if err := bk.applyOptions(opts...); err != nil {
		return nil, err
	}
	id, err := generateRandomID()
	if err != nil {
		return nil, err
	}
	o := &operator{
		id:             id,
		httpRunners:    map[string]*httpRunner{},
		dbRunners:      map[string]*dbRunner{},
		grpcRunners:    map[string]*grpcRunner{},
		cdpRunners:     map[string]*cdpRunner{},
		sshRunners:     map[string]*sshRunner{},
		includeRunners: map[string]*includeRunner{},
		store: store{
			steps:    []map[string]any{},
			stepMap:  map[string]map[string]any{},
			vars:     bk.vars,
			funcs:    bk.funcs,
			bindVars: map[string]any{},
			useMap:   bk.useMap,
		},
		useMap:      bk.useMap,
		desc:        bk.desc,
		labels:      bk.labels,
		debug:       bk.debug,
		profile:     bk.profile,
		interval:    bk.interval,
		loop:        bk.loop,
		concurrency: bk.concurrency,
		t:           bk.t,
		thisT:       bk.t,
		force:       bk.force,
		trace:       bk.trace,
		waitTimeout: bk.waitTimeout,
		failFast:    bk.failFast,
		included:    bk.included,
		ifCond:      bk.ifCond,
		skipTest:    bk.skipTest,
		stdout:      bk.stdout,
		stderr:      bk.stderr,
		newOnly:     bk.loadOnly,
		bookPath:    bk.path,
		beforeFuncs: bk.beforeFuncs,
		afterFuncs:  bk.afterFuncs,
		sw:          stopw.New(),
		capturers:   bk.capturers,
		runResult:   newRunResult(bk.desc, bk.labels, bk.path),
	}

	if o.debug {
		o.capturers = append(o.capturers, NewDebugger(o.stderr))
	}

	root, err := bk.generateOperatorRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to generate root path (%s): %w", bk.path, err)
	}
	o.root = root

	for k, v := range bk.httpRunners {
		if _, ok := v.validator.(*nopValidator); ok {
			for _, l := range bk.openApi3DocLocations {
				key, p := splitKeyAndPath(l)
				if key != "" && key != k {
					continue
				}
				runner, ok := bk.runners[k].(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid type: %v", bk.runners[k])
				}
				c := &httpRunnerConfig{
					OpenApi3DocLocation: p,
				}
				c.SkipValidateRequest, _ = runner["skipValidateRequest"].(bool)
				c.SkipValidateResponse, _ = runner["skipValidateResponse"].(bool)

				val, err := newHttpValidator(c)
				if err != nil {
					return nil, err
				}
				v.validator = val
				break
			}
		}
		if len(bk.hostRules) > 0 {
			v.client.Transport.(*http.Transport).DialContext = bk.hostRules.dialContextFunc()
		}
		o.httpRunners[k] = v
	}
	for k, v := range bk.dbRunners {
		if len(bk.hostRules) > 0 {
			v.hostRules = bk.hostRules
			if err := v.Renew(); err != nil {
				return nil, err
			}
		}
		o.dbRunners[k] = v
	}
	for k, v := range bk.grpcRunners {
		if bk.grpcNoTLS {
			useTLS := false
			v.tls = &useTLS
		}
		for _, proto := range bk.grpcProtos {
			key, p := splitKeyAndPath(proto)
			if key != "" && key != k {
				continue
			}
			v.protos = append(v.protos, p)
		}
		for _, ip := range bk.grpcImportPaths {
			key, p := splitKeyAndPath(ip)
			if key != "" && key != k {
				continue
			}
			v.importPaths = append(v.importPaths, p)
		}
		if len(bk.hostRules) > 0 {
			v.hostRules = bk.hostRules
			if err := v.Renew(); err != nil {
				return nil, err
			}
		}
		o.grpcRunners[k] = v
	}
	for k, v := range bk.cdpRunners {
		if len(bk.hostRules) > 0 {
			v.opts = append(v.opts, bk.hostRules.chromedpOpt())
		}
		if err := v.Renew(); err != nil {
			return nil, err
		}
		o.cdpRunners[k] = v
	}
	for k, v := range bk.sshRunners {
		if len(bk.hostRules) > 0 {
			v.hostRules = bk.hostRules
			if err := v.Renew(); err != nil {
				return nil, err
			}
		}
		o.sshRunners[k] = v
	}
	for k, v := range bk.includeRunners {
		o.includeRunners[k] = v
	}

	keys := map[string]struct{}{}
	for k := range o.httpRunners {
		keys[k] = struct{}{}
	}
	for k := range o.dbRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", o.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	for k := range o.grpcRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", o.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	for k := range o.cdpRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", o.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	for k := range o.sshRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", o.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	for k := range o.includeRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", o.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	var merr error
	for k, err := range bk.runnerErrs {
		merr = multierr.Append(merr, fmt.Errorf("runner %s error: %w", k, err))
	}
	if merr != nil && !o.newOnly {
		return nil, fmt.Errorf("failed to add runners (%s): %w", o.bookPath, merr)
	}

	o.numberOfSteps = len(bk.rawSteps)

	for i, s := range bk.rawSteps {
		key := fmt.Sprintf("%d", i)
		if o.useMap {
			key = bk.stepKeys[i]
		}
		if err := o.appendStep(i, key, s); err != nil {
			if o.newOnly {
				continue
			}
			return nil, fmt.Errorf("failed to append step (%s): %w", o.bookPath, err)
		}
	}

	return o, nil
}

// appendStep appends step.
func (o *operator) appendStep(idx int, key string, s map[string]any) error {
	if o.t != nil {
		o.t.Helper()
	}
	step := newStep(idx, key, o, s)
	// if section
	if v, ok := s[ifSectionKey]; ok {
		step.ifCond, ok = v.(string)
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
	// test runner
	if v, ok := s[testRunnerKey]; ok {
		step.testRunner = newTestRunner()
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
		step.dumpRunner = newDumpRunner()
		switch vv := v.(type) {
		case string:
			step.dumpRequest = &dumpRequest{
				expr: vv,
			}
		case map[string]any:
			expr, ok := vv["expr"]
			if !ok {
				return fmt.Errorf("invalid dump request: %v", vv)
			}
			out, ok := vv["out"]
			if !ok {
				out = "" // default: o.stdout
			}
			step.dumpRequest = &dumpRequest{
				expr: cast.ToString(expr),
				out:  cast.ToString(out),
			}
		default:
			return fmt.Errorf("invalid dump request: %v", vv)
		}
		delete(s, dumpRunnerKey)
	}
	// bind runner
	if v, ok := s[bindRunnerKey]; ok {
		step.bindRunner = newBindRunner()
		cond, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid bind condition: %v", v)
		}
		step.bindCond = cond
		delete(s, bindRunnerKey)
	}

	k, v, ok := pop(s)
	if ok {
		step.runnerKey = k
		switch {
		case k == includeRunnerKey:
			ir, err := newIncludeRunner()
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
			step.execRunner = newExecRunner()
			vv, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid exec command: %v", v)
			}
			step.execCommand = vv
		default:
			detected := false
			h, ok := o.httpRunners[k]
			if ok {
				step.httpRunner = h
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid http request: %v", v)
				}
				step.httpRequest = vv
				detected = true
			}
			db, ok := o.dbRunners[k]
			if ok && !detected {
				step.dbRunner = db
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid db query: %v", v)
				}
				step.dbQuery = vv
				detected = true
			}
			gc, ok := o.grpcRunners[k]
			if ok && !detected {
				step.grpcRunner = gc
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid gRPC request: %v", v)
				}
				step.grpcRequest = vv
				detected = true
			}
			cc, ok := o.cdpRunners[k]
			if ok && !detected {
				step.cdpRunner = cc
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid CDP actions: %v", v)
				}
				step.cdpActions = vv
				detected = true
			}
			sc, ok := o.sshRunners[k]
			if ok && !detected {
				step.sshRunner = sc
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid SSH command: %v", v)
				}
				step.sshCommand = vv
				detected = true
			}
			ic, ok := o.includeRunners[k]
			if ok && !detected {
				step.includeRunner = ic
				c := &includeConfig{
					step: step,
				}
				step.includeConfig = c
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

// Run runbook.
func (o *operator) Run(ctx context.Context) (err error) {
	cctx, cancel := donegroup.WithCancel(ctx)
	defer func() {
		cancel()
		var errr error
		if o.waitTimeout > 0 {
			errr = donegroup.WaitWithTimeout(cctx, o.waitTimeout)
		} else {
			errr = donegroup.Wait(cctx)
		}
		err = errors.Join(err, errr)
	}()
	if o.t != nil {
		o.t.Helper()
	}
	if !o.profile {
		o.sw.Disable()
	}
	defer o.sw.Start().Stop()
	defer func() {
		o.capturers.captureResult(o.trails(), o.Result())
		o.capturers.captureEnd(o.trails(), o.bookPath, o.desc)
		o.Close(true)
	}()
	o.capturers.captureStart(o.trails(), o.bookPath, o.desc)
	if err := o.run(cctx); err != nil {
		return err
	}
	return nil
}

// DumpProfile write run time profile.
func (o *operator) DumpProfile(w io.Writer) error {
	r := o.sw.Result()
	if r == nil {
		return errors.New("no profile")
	}
	// Use encoding/json because goccy/go-json got a SIGSEGV error due to the increase in Trail fields.
	enc := ejson.NewEncoder(w)
	if err := enc.Encode(r); err != nil {
		return err
	}
	return nil
}

// Result returns run result.
func (o *operator) Result() *RunResult {
	o.runResult.ID = o.runbookID()
	r := o.sw.Result()
	if r != nil {
		if err := setElasped(o.runResult, r); err != nil {
			panic(err)
		}
	}
	return o.runResult
}

func (o *operator) clearResult() {
	o.runResult = newRunResult(o.desc, o.labels, o.bookPathOrID())
	o.runResult.ID = o.runbookID()
	for _, s := range o.steps {
		s.clearResult()
	}
}

func (o *operator) run(ctx context.Context) error {
	defer o.sw.Start(o.trails().toProfileIDs()...).Stop()
	if o.newOnly {
		return errors.New("this runbook is not allowed to run")
	}
	var err error
	if o.t != nil {
		// As test helper
		o.t.Helper()
		o.t.Run(o.testName(), func(t *testing.T) {
			t.Helper()
			o.thisT = t
			if o.loop != nil {
				err = o.runLoop(ctx)
			} else {
				err = o.runInternal(ctx)
			}
			if err != nil {
				// Skip parent runner t.Error if there is an error in the included runbook
				if !errors.Is(&includedRunErr{}, err) {
					paths, indexes, errs := failedRunbookPathsAndErrors(o.runResult)
					for ii, p := range paths {
						last := p[len(p)-1]
						b, err := readFile(last)
						if err != nil {
							t.Error(errs[ii])
							continue
						}
						idx := indexes[ii]
						var fs string
						if idx >= 0 {
							picked, err := pickStepYAML(string(b), idx)
							if err != nil {
								t.Error(errs[ii])
								continue
							}
							fs = fmt.Sprintf("Failure step (%s):\n%s\n\n", last, picked)
						}
						if !strings.HasSuffix(errs[ii].Error(), "\n") {
							fs = "\n" + fs
						}
						t.Errorf("%s%s\n", red(errs[ii]), fs)
					}
				}
			}
		})
		o.thisT = o.t
		if err != nil {
			return fmt.Errorf("failed to run %s: %w", o.bookPathOrID(), err)
		}
		return nil
	}
	if o.loop != nil {
		err = o.runLoop(ctx)
	} else {
		err = o.runInternal(ctx)
	}
	if err != nil {
		return fmt.Errorf("failed to run %s: %w", o.bookPathOrID(), err)
	}
	return nil
}

func (o *operator) runLoop(ctx context.Context) error {
	if o.loop == nil {
		panic("invalid usage")
	}
	retrySuccess := false
	if o.loop.Until == "" {
		retrySuccess = true
	}
	var (
		err     error
		outcome result
		bt      string
		j       int
	)
	c, err := EvalCount(o.loop.Count, o.store.toMap())
	if err != nil {
		return err
	}
	var looperr error
	for o.loop.Loop(ctx) {
		if j >= c {
			break
		}
		if j > 0 {
			// Renew runners
			for _, r := range o.cdpRunners {
				if err := r.Renew(); err != nil {
					return err
				}
			}
		}
		i := j
		o.loopIndex = &i
		trs := o.trails()
		o.capturers.setCurrentTrails(trs)
		sw := o.sw.Start(trs.toProfileIDs()...)
		err = o.runInternal(ctx)
		if err != nil {
			sw.Stop()
			looperr = multierr.Append(looperr, fmt.Errorf("loop[%d]: %w", j, err))
			outcome = resultFailure
		} else {
			sw.Stop()
			if o.Skipped() {
				outcome = resultSkipped
			} else {
				outcome = resultSuccess
			}
		}
		if o.loop.Until != "" {
			store := o.store.toMap()
			store[storeStepKeyOutcome] = string(outcome)
			bt, err = buildTree(o.loop.Until, store)
			if err != nil {
				return fmt.Errorf("loop failed on %s: %w", o.bookPathOrID(), err)
			}
			tf, err := EvalCond(o.loop.Until, store)
			if err != nil {
				return fmt.Errorf("loop failed on %s: %w", o.bookPathOrID(), err)
			}
			if tf {
				retrySuccess = true
				break
			}
		}
		j++
	}
	if !retrySuccess {
		err := fmt.Errorf("(%s) is not true\n%s", o.loop.Until, bt)
		if o.loop.interval != nil {
			return fmt.Errorf("retry loop failed on %s.loop (count: %d, interval: %v): %w", o.bookPathOrID(), c, *o.loop.interval, err)
		} else {
			return fmt.Errorf("retry loop failed on %s.loop (count: %d, minInterval: %v, maxInterval: %v): %w", o.bookPathOrID(), c, *o.loop.minInterval, *o.loop.maxInterval, err)
		}
	}
	if o.loop.Until == "" && looperr != nil {
		// simple count
		return fmt.Errorf("loop failed on %s: %w", o.bookPathOrID(), looperr)
	}

	return nil
}

func (o *operator) runInternal(ctx context.Context) (rerr error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.t != nil {
		o.t.Helper()
	}
	// Clear results for each scenario run (runInternal); results per root loop are not retrievable.
	o.clearResult()
	o.store.clearSteps()

	defer func() {
		// Set run error and skipped status
		o.runResult.Err = rerr
		o.runResult.Skipped = o.Skipped()
		o.runResult.Store = o.store.toMap()
		o.runResult.StepResults = o.StepResults()

		if o.Skipped() {
			// If the scenario is skipped, beforeFuncs/afterFuncs are not executed
			return
		}

		// afterFuncs
		for i, fn := range o.afterFuncs {
			i := i
			trs := append(o.trails(), Trail{
				Type:      TrailTypeAfterFunc,
				FuncIndex: &i,
			})
			trsi := trs.toProfileIDs()
			o.sw.Start(trsi...)
			if aferr := fn(o.runResult); aferr != nil {
				rerr = newAfterFuncError(aferr)
				o.runResult.Err = rerr
			}
			o.sw.Stop(trsi...)
		}
	}()

	// if
	if o.ifCond != "" {
		tf, err := o.expandCondBeforeRecord(o.ifCond)
		if err != nil {
			rerr = err
			return
		}
		if !tf {
			if err := o.skip(); err != nil {
				rerr = err
				return
			}
			return nil
		}
	}

	// beforeFuncs
	o.runResult.Store = o.store.toMap()
	for i, fn := range o.beforeFuncs {
		i := i
		trs := append(o.trails(), Trail{
			Type:      TrailTypeBeforeFunc,
			FuncIndex: &i,
		})
		trsi := trs.toProfileIDs()
		o.sw.Start(trsi...)
		if err := fn(o.runResult); err != nil {
			o.sw.Stop(trsi...)
			return newBeforeFuncError(err)
		}
		o.sw.Stop(trsi...)
	}

	// steps
	failed := false
	force := o.force
	for i, s := range o.steps {
		if failed && !force {
			s.setResult(errStepSkiped)
			o.recordNotRun(i)
			if err := o.recordToLatest(storeStepKeyOutcome, resultSkipped); err != nil {
				return err
			}
			continue
		}
		err := o.runStep(ctx, i, s)
		s.setResult(err)
		switch {
		case errors.Is(errStepSkiped, err):
			o.recordNotRun(i)
			if err := o.recordToLatest(storeStepKeyOutcome, resultSkipped); err != nil {
				return err
			}
		case err != nil:
			o.recordNotRun(i)
			if err := o.recordToLatest(storeStepKeyOutcome, resultFailure); err != nil {
				return err
			}
			rerr = multierr.Append(rerr, err)
			failed = true
		default:
			if err := o.recordToLatest(storeStepKeyOutcome, resultSuccess); err != nil {
				return err
			}
		}
	}

	return
}

func (o *operator) bookPathOrID() string {
	if o.bookPath != "" {
		return o.bookPath
	}
	return o.id
}

func (o *operator) testName() string {
	if o.bookPath == "" {
		return fmt.Sprintf("-(%s)", o.id)
	}
	return fmt.Sprintf("%s(%s)", o.bookPath, o.id)
}

func (o *operator) stepName(i int) string {
	var prefix string

	if o.store.loopIndex != nil {
		prefix = fmt.Sprintf(".loop[%d]", *o.store.loopIndex)
	}
	if o.useMap {
		return fmt.Sprintf("%q.steps.%s%s", o.desc, o.steps[i].key, prefix)
	}

	return fmt.Sprintf("%q.steps[%d]%s", o.desc, i, prefix)
}

// expandBeforeRecord - expand before the runner records the result.
func (o *operator) expandBeforeRecord(in any) (any, error) {
	store := o.store.toMap()
	store[storeRootKeyIncluded] = o.included
	store[storeRootPrevious] = o.store.latest()
	return EvalExpand(in, store)
}

// expandCondBeforeRecord - expand condition before the runner records the result.
func (o *operator) expandCondBeforeRecord(ifCond string) (bool, error) {
	store := o.store.toMap()
	store[storeRootKeyIncluded] = o.included
	store[storeRootPrevious] = o.store.latest()
	return EvalCond(ifCond, store)
}

// Debugln print to out when debug = true.
func (o *operator) Debugln(a any) {
	if !o.debug {
		return
	}
	_, _ = fmt.Fprintln(o.stderr, a)
}

// Debugf print to out when debug = true.
func (o *operator) Debugf(format string, a ...any) {
	if !o.debug {
		return
	}
	_, _ = fmt.Fprintf(o.stderr, format, a...)
}

// Warnf print to out.
func (o *operator) Warnf(format string, a ...any) {
	_, _ = fmt.Fprintf(o.stderr, format, a...)
}

// Skipped returns whether the runbook run skipped.
func (o *operator) Skipped() bool {
	return o.skipped
}

func (o *operator) skip() error {
	o.Debugf(yellow("Skip %s\n"), o.desc)
	o.skipped = true
	for i, s := range o.steps {
		s.setResult(errStepSkiped)
		o.recordNotRun(i)
		if err := o.recordToLatest(storeStepKeyOutcome, resultSkipped); err != nil {
			return err
		}
	}
	return nil
}

func (o *operator) StepResults() []*StepResult {
	var results []*StepResult
	for _, s := range o.steps {
		results = append(results, s.result)
	}
	return results
}

type operators struct {
	ops         []*operator
	t           *testing.T
	sw          *stopw.Span
	profile     bool
	shuffle     bool
	shuffleSeed int64
	shardN      int
	shardIndex  int
	sample      int
	random      int
	waitTimeout time.Duration // waitTimout is the time to wait for sub-processes to complete after the Run or RunN context is canceled.
	concmax     int
	opts        []Option
	results     []*runNResult
	runCount    int64
	mu          sync.Mutex
}

func Load(pathp string, opts ...Option) (*operators, error) {
	bk := newBook()
	envOpts := []Option{
		RunMatch(os.Getenv("RUNN_RUN")),
		RunID(os.Getenv("RUNN_ID")),
		RunLabel(os.Getenv("RUNN_LABEL")),
		Scopes(os.Getenv("RUNN_SCOPES")),
	}
	opts = append(envOpts, opts...)
	if err := bk.applyOptions(opts...); err != nil {
		return nil, err
	}

	sw := stopw.New()
	ops := &operators{
		t:           bk.t,
		sw:          sw,
		profile:     bk.profile,
		shuffle:     bk.runShuffle,
		shuffleSeed: bk.runShuffleSeed,
		shardN:      bk.runShardN,
		shardIndex:  bk.runShardIndex,
		sample:      bk.runSample,
		random:      bk.runRandom,
		waitTimeout: bk.waitTimeout,
		concmax:     1,
		opts:        opts,
	}
	if bk.runConcurrent {
		ops.concmax = bk.runConcurrentMax
	}
	books, err := Books(pathp)
	if err != nil {
		return nil, err
	}
	var skipPaths []string
	om := map[string]*operator{}
	var opss []*operator
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
		opss = append(opss, o)
	}

	if err := generateIDsUsingPath(opss); err != nil {
		return nil, err
	}

	var idMatched []*operator
	cond := labelCond(bk.runLabels)
	indexes := map[string]int{}
	for p, o := range om {
		// RUNN_RUN, --run
		if !bk.runMatch.MatchString(p) {
			o.Debugf(yellow("Skip %s because it does not match %s\n"), p, bk.runMatch.String())
			continue
		}
		if contains(skipPaths, p) {
			o.Debugf(yellow("Skip %s because it is already included from another runbook\n"), p)
			continue
		}
		// RUUN_LABEL, --label
		tf, err := EvalCond(cond, labelEnv(o.labels))
		if err != nil {
			return nil, err
		}
		if !tf {
			o.Debugf(yellow("Skip %s because it does not match %s\n"), p, cond)
			continue
		}
		// RUUN_ID, --id
		for i, id := range bk.runIDs {
			if strings.HasPrefix(o.id, id) {
				idMatched = append(idMatched, o)
				indexes[o.id] = i
			}
		}
		o.sw = ops.sw
		ops.ops = append(ops.ops, o)
	}

	// Run the matching runbooks in order if there is only one runbook with a forward matching ID.
	if len(bk.runIDs) > 0 {
		switch {
		case len(idMatched) == 0:
			return nil, fmt.Errorf("no runbooks has the id prefix: %s", bk.runIDs)
		default:
			u := lo.UniqBy(idMatched, func(o *operator) string {
				return o.id
			})
			if len(u) != len(idMatched) {
				return nil, fmt.Errorf("multiple runbooks have the same id prefix: %s", bk.runIDs)
			}
			// Sort the matching runbooks in the order of the specified IDs.
			sort.SliceStable(idMatched, func(i, j int) bool {
				ii, ok := indexes[idMatched[i].id]
				if !ok {
					return false
				}
				jj, ok := indexes[idMatched[j].id]
				if !ok {
					return false
				}
				return ii < jj
			})
			ops.ops = idMatched
		}
	} else {
		// If no ids are specified, the order is sorted and fixed
		sortOperators(ops.ops)
	}
	return ops, nil
}

func (ops *operators) RunN(ctx context.Context) (err error) {
	cctx, cancel := donegroup.WithCancel(ctx)
	defer func() {
		cancel()
		var errr error
		if ops.waitTimeout > 0 {
			errr = donegroup.WaitWithTimeout(cctx, ops.waitTimeout)
		} else {
			errr = donegroup.Wait(cctx)
		}
		err = errors.Join(err, errr)
	}()
	if ops.t != nil {
		ops.t.Helper()
	}
	result, err := ops.runN(cctx)
	ops.mu.Lock()
	ops.results = append(ops.results, result)
	ops.mu.Unlock()
	if err != nil {
		return err
	}
	return nil
}

func (ops *operators) Operators() []*operator {
	return ops.ops
}

func (ops *operators) Close() {
	for _, o := range ops.ops {
		o.Close(true)
	}
}

func (ops *operators) DumpProfile(w io.Writer) error {
	r := ops.sw.Result()
	if r == nil {
		return errors.New("no profile")
	}
	enc := ejson.NewEncoder(w)
	if err := enc.Encode(r); err != nil {
		return err
	}
	return nil
}

func (ops *operators) Init() error {
	return nil
}

func (ops *operators) RequestOne(ctx context.Context) error {
	result, err := ops.runN(ctx)
	if err != nil {
		return err
	}
	if result.HasFailure() {
		return errors.New("result has failure")
	}
	return nil
}

func (ops *operators) Terminate() error {
	ops.Close()
	return nil
}

func (ops *operators) Result() *runNResult {
	return ops.results[len(ops.results)-1]
}

func (ops *operators) SelectedOperators() ([]*operator, error) {
	var err error
	rc := ops.runCount
	atomic.AddInt64(&ops.runCount, 1)
	tops := make([]*operator, len(ops.ops))
	copy(tops, ops.ops)
	if rc > 0 && ops.random == 0 {
		tops, err = copyOperators(tops, ops.opts)
		if err != nil {
			return nil, err
		}
	}
	if ops.shuffle {
		// Shuffle order of running
		shuffleOperators(tops, ops.shuffleSeed)
	}

	if ops.shardN > 0 {
		tops = partOperators(tops, ops.shardN, ops.shardIndex)
	}
	if ops.sample > 0 {
		tops = sampleOperators(tops, ops.sample)
	}
	if ops.random > 0 {
		rops, err := randomOperators(tops, ops.opts, ops.random)
		if err != nil {
			return nil, err
		}
		for _, o := range rops {
			o.sw = ops.sw
		}
		return rops, nil
	}

	return tops, nil
}

func (ops *operators) CollectCoverage(ctx context.Context) (*Coverage, error) {
	cov := &Coverage{}
	for _, o := range ops.ops {
		c, err := o.collectCoverage(ctx)
		if err != nil {
			return nil, err
		}
		// Merge coverage
		for _, sc := range c.Specs {
			spec, ok := lo.Find(cov.Specs, func(i *SpecCoverage) bool {
				return sc.Key == i.Key
			})
			if !ok {
				cov.Specs = append(cov.Specs, sc)
				continue
			}
			for k, v := range sc.Coverages {
				spec.Coverages[k] += v
			}
		}
	}
	sort.SliceStable(cov.Specs, func(i, j int) bool {
		return cov.Specs[i].Key < cov.Specs[j].Key
	})
	return cov, nil
}

func (ops *operators) runN(ctx context.Context) (*runNResult, error) {
	result := &runNResult{}
	if ops.t != nil {
		ops.t.Helper()
	}
	if !ops.profile {
		ops.sw.Disable()
	}
	defer ops.sw.Start().Stop()
	defer ops.Close()
	cg, cctx := concgroup.WithContext(ctx)
	cg.SetLimit(ops.concmax)
	selected, err := ops.SelectedOperators()
	if err != nil {
		return result, err
	}
	result.Total.Add(int64(len(selected)))
	for _, o := range selected {
		o := o
		cg.GoMulti(o.concurrency, func() error {
			select {
			case <-cctx.Done():
				return errors.New("context canceled")
			default:
			}
			defer func() {
				r := o.Result()
				o.capturers.captureResult(o.trails(), r)
				o.capturers.captureEnd(o.trails(), o.bookPath, o.desc)
				o.Close(false)
				result.mu.Lock()
				result.RunResults = append(result.RunResults, r)
				result.mu.Unlock()
			}()
			o.capturers.captureStart(o.trails(), o.bookPath, o.desc)
			if err := o.run(cctx); err != nil {
				if o.failFast {
					return err
				}
			}
			return nil
		})
	}
	if err := cg.Wait(); err != nil {
		return result, err
	}
	return result, nil
}

func partOperators(ops []*operator, n, i int) []*operator {
	all := make([]*operator, len(ops))
	copy(all, ops)
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

func copyOperators(ops []*operator, opts []Option) ([]*operator, error) {
	var c []*operator
	for _, o := range ops {
		// FIXME: Need the function to copy the operator as it is heavy to parse the runbook each time
		oo, err := New(append([]Option{Book(o.bookPath)}, opts...)...)
		if err != nil {
			return nil, err
		}
		oo.id = o.id // Copy id from original operator
		c = append(c, oo)
	}
	return c, nil
}

func sampleOperators(ops []*operator, num int) []*operator {
	if len(ops) <= num {
		return ops
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	var sample []*operator
	n := make([]*operator, len(ops))
	copy(n, ops)

	for i := 0; i < num; i++ {
		idx := r.Intn(len(n))
		sample = append(sample, n[idx])
		n = append(n[:idx], n[idx+1:]...)
	}
	return sample
}

func randomOperators(ops []*operator, opts []Option, num int) ([]*operator, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	var random []*operator
	n := make([]*operator, len(ops))
	copy(n, ops)
	for i := 0; i < num; i++ {
		idx := r.Intn(len(n))
		// FIXME: Need the function to copy the operator as it is heavy to parse the runbook each time
		o, err := New(append([]Option{Book(n[idx].bookPath)}, opts...)...)
		if err != nil {
			return nil, err
		}
		random = append(random, o)
	}
	return random, nil
}

func shuffleOperators(ops []*operator, seed int64) {
	r := rand.New(rand.NewSource(seed)) //nolint:gosec
	r.Shuffle(len(ops), func(i, j int) {
		ops[i], ops[j] = ops[j], ops[i]
	})
}

func pop(s map[string]any) (string, any, bool) {
	for k, v := range s {
		defer delete(s, k)
		return k, v, true
	}
	return "", nil, false
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}

func setElasped(r *RunResult, result *stopw.Span) error {
	m := collectStepElaspedByRunbookIDFull(result, nil, map[string]time.Duration{})
	return setElaspedByRunbookIDFull(r, m)
}

// collectStepElaspedByRunbookIDFull collects the elapsed time of each step by runbook ID.
func collectStepElaspedByRunbookIDFull(r *stopw.Span, trs Trails, m map[string]time.Duration) map[string]time.Duration {
	var t Trail
	t, ok := r.ID.(Trail)
	if ok {
		trs = append(trs, t)
		switch t.Type {
		case TrailTypeRunbook:
			id := trs.runbookID()
			if !strings.Contains(id, "?step=") {
				// Collect root runbook only
				m[id] += r.Elapsed()
			}
		case TrailTypeStep:
			// Collect steps
			id := trs.runbookID()
			m[id] += r.Elapsed()
		}
	}
	for _, b := range r.Breakdown {
		m = collectStepElaspedByRunbookIDFull(b, trs, m)
	}
	return m
}

// setElaspedByRunbookIDFull sets the elapsed time.
func setElaspedByRunbookIDFull(r *RunResult, m map[string]time.Duration) error {
	e, ok := m[r.ID]
	if !ok {
		return nil
	}
	r.Elapsed = e
	for _, sr := range r.StepResults {
		if sr == nil {
			continue
		}
		e, ok := m[sr.ID]
		if !ok {
			continue
		}
		sr.Elapsed = e
		if sr.IncludedRunResult != nil {
			if err := setElaspedByRunbookIDFull(sr.IncludedRunResult, m); err != nil {
				return err
			}
		}
	}
	return nil
}

var labelRep = strings.NewReplacer("-", "___hyphen___", "/", "___slash___", ".", "___dot___", ":", "___colon___")

func labelEnv(labels []string) map[string]bool {
	return lo.SliceToMap(labels, func(l string) (string, bool) {
		return labelRep.Replace(l), true
	})
}

func labelCond(labels []string) string {
	if len(labels) == 0 {
		return "true"
	}
	return labelRep.Replace("(" + strings.Join(labels, ") or (") + ")")
}
