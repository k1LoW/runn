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
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/k1LoW/concgroup"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/maskedio"
	"github.com/k1LoW/runn/internal/deprecation"
	"github.com/k1LoW/runn/internal/exprtrace"
	"github.com/k1LoW/runn/internal/kv"
	"github.com/k1LoW/stopw"
	"github.com/k1LoW/waitmap"
	"github.com/ryo-yamaoka/otchkiss"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

var errStepSkipped = errors.New("step skipped")
var ErrFailFast = errors.New("fail fast")

var _ otchkiss.Requester = (*operatorN)(nil)

type need struct {
	path string
	op   *operator
}

type deferredOpAndStep struct {
	op   *operator
	step *step
}

type deferredOpAndSteps struct {
	steps []*deferredOpAndStep
}

type operator struct {
	id              string
	httpRunners     map[string]*httpRunner
	dbRunners       map[string]*dbRunner
	grpcRunners     map[string]*grpcRunner
	cdpRunners      map[string]*cdpRunner
	sshRunners      map[string]*sshRunner
	includeRunners  map[string]*includeRunner
	steps           []*step
	deferred        *deferredOpAndSteps
	store           *store
	desc            string
	needs           map[string]*need                 // Map of `needs:` in runbook. key is the operator.bookPath.
	nm              *waitmap.WaitMap[string, *store] // Map of runbook result stores. key is the operator.bookPath.
	labels          []string
	useMap          bool // Use map syntax in `steps:`.
	debug           bool // Enable debug mode
	profile         bool
	interval        time.Duration
	loop            *Loop
	loopIndex       *int // Index of the loop is dynamically recorded at runtime
	concurrency     []string
	root            string // Root directory of runbook ( rubbook path or working directory )
	t               *testing.T
	thisT           *testing.T
	parent          *step
	force           bool
	trace           bool // Enable tracing ( e.g. add trace header to HTTP request )
	waitTimeout     time.Duration
	included        bool
	ifCond          string
	skipTest        bool
	skipped         bool
	stdout          *maskedio.Writer
	stderr          *maskedio.Writer
	newOnly         bool // Skip some errors for `runn list`
	bookPath        string
	numberOfSteps   int // Number of steps for `runn list`
	beforeFuncs     []func(*RunResult) error
	afterFuncs      []func(*RunResult) error
	sw              *stopw.Span
	capturers       capturers
	runResult       *RunResult
	dbg             *dbg
	hasRunnerRunner bool
	maskRule        *maskedio.Rule

	mu sync.Mutex
}

// ID returns id of current runbook.
func (op *operator) ID() string {
	return op.id
}

// runbookID returns id of the root runbook.
func (op *operator) runbookID() string { //nolint:unused
	return op.trails().runbookID()
}

// Desc returns `desc:` of runbook.
func (op *operator) Desc() string {
	return op.desc
}

// If returns `if:` of runbook.
func (op *operator) If() string {
	return op.ifCond
}

// BookPath returns path of runbook.
func (op *operator) BookPath() string {
	return op.bookPath
}

// NumberOfSteps returns number of steps.
func (op *operator) NumberOfSteps() int {
	return op.numberOfSteps
}

// Store returns stored values.
// Deprecated: Use Result().Store() instead.
func (op *operator) Store() map[string]any {
	deprecation.AddWarning("operator.Store", "Use Result().Store() instead.")
	return op.Result().Store()
}

// Close runners.
func (op *operator) Close(force bool) {
	for _, r := range op.grpcRunners {
		if !force && r.target == "" {
			continue
		}
		_ = r.Close()
	}
	for _, r := range op.cdpRunners {
		_ = r.Close()
	}
	for _, r := range op.sshRunners {
		_ = r.Close()
	}
	for _, r := range op.dbRunners {
		if !force && r.dsn == "" {
			continue
		}
		_ = r.Close()
	}
}

func (op *operator) runStep(ctx context.Context, s *step) error {
	idx := s.idx
	if op.t != nil {
		op.t.Helper()
	}
	if err := op.dbg.attach(ctx, s); err != nil {
		return err
	}
	trs := s.trails()
	defer op.sw.Start(trs.toProfileIDs()...).Stop()
	op.capturers.setCurrentTrails(trs)
	if idx != 0 {
		// interval:
		time.Sleep(op.interval)
		op.Debugln("")
	}
	if s.ifCond != "" {
		tf, err := op.expandCondBeforeRecord(s.ifCond, s)
		if err != nil {
			return err
		}
		if !tf {
			if s.desc != "" {
				op.Debugf(yellow("Skip %q on %s\n"), s.desc, op.stepName(idx))
			} else if s.runnerKey != "" {
				op.Debugf(yellow("Skip %q on %s\n"), s.runnerKey, op.stepName(idx))
			} else {
				op.Debugf(yellow("Skip on %s\n"), op.stepName(idx))
			}
			return errStepSkipped
		}
	}
	if s.desc != "" {
		op.Debugf(cyan("Run %q on %s\n"), s.desc, op.stepName(idx))
	} else if s.runnerKey != "" {
		op.Debugf(cyan("Run %q on %s\n"), s.runnerKey, op.stepName(idx))
	}

	stepFn := func(t *testing.T) error {
		s.clearResult()
		if t != nil {
			t.Helper()
		}
		run := false
		if s.notYetDetectedRunner() {
			if r, ok := op.httpRunners[s.runnerKey]; ok {
				s.httpRunner = r
				s.httpRequest = s.runnerValues
			}
			if r, ok := op.dbRunners[s.runnerKey]; ok {
				s.dbRunner = r
				s.dbQuery = s.runnerValues
			}
			if r, ok := op.grpcRunners[s.runnerKey]; ok {
				s.grpcRunner = r
				s.grpcRequest = s.runnerValues
			}
			if r, ok := op.cdpRunners[s.runnerKey]; ok {
				s.cdpRunner = r
				s.cdpActions = s.runnerValues
			}
			if r, ok := op.sshRunners[s.runnerKey]; ok {
				s.sshRunner = r
				s.sshCommand = s.runnerValues
			}
		}
		switch {
		case s.httpRunner != nil && s.httpRequest != nil:
			if err := s.httpRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("http request failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		case s.dbRunner != nil && s.dbQuery != nil:
			if err := s.dbRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("db query failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		case s.grpcRunner != nil && s.grpcRequest != nil:
			if err := s.grpcRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("gRPC request failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		case s.cdpRunner != nil && s.cdpActions != nil:
			if err := s.cdpRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("cdp action failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		case s.sshRunner != nil && s.sshCommand != nil:
			if err := s.sshRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("ssh command failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		case s.execRunner != nil && s.execCommand != nil:
			if err := s.execRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("exec command failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		case s.includeRunner != nil && s.includeConfig != nil:
			if err := s.includeRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("include failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		case s.runnerRunner != nil && s.runnerDefinition != nil:
			if err := s.runnerRunner.Run(ctx, s); err != nil {
				return fmt.Errorf("runner definition failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		}
		// dump runner
		if s.dumpRunner != nil && s.dumpRequest != nil {
			op.Debugf(cyan("Run %q on %s\n"), dumpRunnerKey, op.stepName(idx))
			if err := s.dumpRunner.Run(ctx, s, !run); err != nil {
				return fmt.Errorf("dump failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		}
		// bind runner
		if s.bindRunner != nil && s.bindCond != nil {
			op.Debugf(cyan("Run %q on %s\n"), bindRunnerKey, op.stepName(idx))
			if err := s.bindRunner.Run(ctx, s, !run); err != nil {
				return fmt.Errorf("bind failed on %s: %w", op.stepName(idx), err)
			}
			run = true
		}
		// test runner
		if s.testRunner != nil && s.testCond != "" {
			if op.skipTest {
				op.Debugf(yellow("Skip %q on %s\n"), testRunnerKey, op.stepName(idx))
				if !run {
					return errStepSkipped
				}
				return nil
			}
			op.Debugf(cyan("Run %q on %s\n"), testRunnerKey, op.stepName(idx))
			if err := s.testRunner.Run(ctx, s, !run); err != nil {
				if s.desc != "" {
					return fmt.Errorf("test failed on %s %q: %w", op.stepName(idx), s.desc, err)
				} else {
					return fmt.Errorf("test failed on %s: %w", op.stepName(idx), err)
				}
			}
			run = true
		}

		if !run {
			return fmt.Errorf("invalid runner: %v", op.stepName(idx))
		}
		return nil
	}

	// loop
	if s.loop != nil {
		defer func() {
			op.store.loopIndex = nil
			s.loopIndex = nil
			s.loop.Clear()
		}()
		retrySuccess := false
		if s.loop.Until == "" {
			retrySuccess = true
		}
		var (
			bt string
			j  int
		)
		c, err := EvalCount(s.loop.Count, op.store.toMap())
		if err != nil {
			return err
		}
		for s.loop.Loop(ctx) {
			if j >= c {
				break
			}
			jj := j
			op.store.loopIndex = &jj
			s.loopIndex = &jj
			trs := s.trails()
			op.capturers.setCurrentTrails(trs)
			sw := op.sw.Start(trs.toProfileIDs()...)
			if err := stepFn(op.thisT); err != nil {
				sw.Stop()
				return fmt.Errorf("loop failed: %w", err)
			}
			sw.Stop()
			if s.loop.Until != "" {
				store := op.store.toMap()
				store[storeRootKeyIncluded] = op.included
				if !s.deferred {
					store[storeRootKeyPrevious] = op.store.previous()
				}
				store[storeRootKeyCurrent] = op.store.latest()
				tf, err := EvalWithTrace(s.loop.Until, store)
				if err != nil {
					return fmt.Errorf("loop failed on %s: %w", op.stepName(idx), err)
				}
				if tf.OutputAsBool() {
					retrySuccess = true
					break
				} else {
					bt, err = tf.FormatTraceTree()
					if err != nil {
						return fmt.Errorf("loop failed on %s: %w", op.stepName(idx), err)
					}
				}
			}
			j++
		}
		if !retrySuccess {
			err := fmt.Errorf("(%s) is not true\n%s", s.loop.Until, bt)
			if s.loop.interval != nil {
				return fmt.Errorf("retry loop failed on %s.loop (count: %d, interval: %v): %w", op.stepName(idx), c, *s.loop.interval, err)
			} else {
				return fmt.Errorf("retry loop failed on %s.loop (count: %d, minInterval: %v, maxInterval: %v): %w", op.stepName(idx), c, *s.loop.minInterval, *s.loop.maxInterval, err)
			}
		}
	} else {
		if err := stepFn(op.thisT); err != nil {
			return err
		}
	}
	return nil
}

// Record that it has not been run.
func (op *operator) recordNotRun(idx int) {
	v := map[string]any{}
	op.store.record(idx, v)
}

func (op *operator) record(idx int, v map[string]any) {
	if v == nil {
		v = map[string]any{}
	}
	op.store.record(idx, v)
}

func (op *operator) recordResult(idx int, v result) error {
	r := op.Result()
	r.StepResults = op.StepResults()
	op.capturers.captureResultByStep(op.trails(), r)
	return op.store.recordTo(idx, storeStepKeyOutcome, v)
}

func (op *operator) recordCookie(cookies []*http.Cookie) {
	op.store.recordCookie(cookies)
}

func (op *operator) generateTrail() Trail {
	return Trail{
		Type:        TrailTypeRunbook,
		Desc:        op.desc,
		RunbookID:   op.id,
		RunbookPath: op.bookPath,
	}
}

func (op *operator) trails() Trails {
	var trs Trails
	if op.parent != nil {
		trs = op.parent.trails()
	}
	trs = append(trs, op.generateTrail())
	if op.loopIndex != nil {
		trs = append(trs, Trail{
			Type:      TrailTypeLoop,
			LoopIndex: op.loopIndex,
			RunbookID: op.id,
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
	st := newStore(bk.vars, bk.funcs, bk.secrets, bk.useMap, bk.stepKeys)
	op := &operator{
		id:             id,
		httpRunners:    map[string]*httpRunner{},
		dbRunners:      map[string]*dbRunner{},
		grpcRunners:    map[string]*grpcRunner{},
		cdpRunners:     map[string]*cdpRunner{},
		sshRunners:     map[string]*sshRunner{},
		includeRunners: map[string]*includeRunner{},
		deferred:       &deferredOpAndSteps{},
		store:          st,
		useMap:         bk.useMap,
		desc:           bk.desc,
		labels:         bk.labels,
		debug:          bk.debug,
		nm:             waitmap.New[string, *store](),
		profile:        bk.profile,
		interval:       bk.interval,
		loop:           bk.loop,
		concurrency:    bk.concurrency,
		t:              bk.t,
		thisT:          bk.t,
		force:          bk.force,
		trace:          bk.trace,
		waitTimeout:    bk.waitTimeout,
		included:       bk.included,
		ifCond:         bk.ifCond,
		skipTest:       bk.skipTest,
		stdout:         st.maskRule().NewWriter(bk.stdout),
		stderr:         st.maskRule().NewWriter(bk.stderr),
		newOnly:        bk.loadOnly,
		bookPath:       bk.path,
		beforeFuncs:    bk.beforeFuncs,
		afterFuncs:     bk.afterFuncs,
		sw:             stopw.New(),
		capturers:      bk.capturers,
		runResult:      newRunResult(bk.desc, bk.labels, bk.path, bk.included, st),
		dbg:            newDBG(bk.attach),
		maskRule:       st.maskRule(),
	}

	if op.debug {
		op.capturers = append(op.capturers, NewDebugger(op.stderr))
	}

	root, err := bk.generateOperatorRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to generate root path (%s): %w", bk.path, err)
	}
	op.root = root

	var loErr error
	op.needs = lo.MapEntries(bk.needs, func(key string, path string) (string, *need) {
		p, err := fp(path, op.root)
		if err != nil {
			loErr = errors.Join(loErr, err)
		}
		return key, &need{
			path: p,
		}
	})
	if loErr != nil {
		return nil, loErr
	}

	// The host rules specified by the option take precedence.
	hostRules := append(bk.hostRulesFromOpts, bk.hostRules...)

	for k, v := range bk.httpRunners {
		if _, ok := v.validator.(*nopValidator); ok {
			for _, l := range bk.openAPI3DocLocations {
				key, p := splitKeyAndPath(l)
				if key != "" && key != k {
					continue
				}
				runner, ok := bk.runners[k].(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid type: %v", bk.runners[k])
				}
				c := &httpRunnerConfig{
					OpenAPI3DocLocation: p,
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
		if len(hostRules) > 0 {
			tp, ok := v.client.Transport.(*http.Transport)
			if !ok {
				return nil, fmt.Errorf("failed to cast: %v", v.client.Transport)
			}
			tp.DialContext = hostRules.dialContextFunc()
		}
		op.httpRunners[k] = v
	}
	for k, v := range bk.dbRunners {
		if len(hostRules) > 0 {
			v.hostRules = hostRules
			if err := v.Renew(); err != nil {
				return nil, err
			}
		}
		if v.operatorID == "" {
			v.operatorID = op.id
		}
		op.dbRunners[k] = v
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
		v.bufDirs = unique(append(v.bufDirs, bk.grpcBufDirs...))
		v.bufLocks = unique(append(v.bufLocks, bk.grpcBufLocks...))
		v.bufConfigs = unique(append(v.bufConfigs, bk.grpcBufConfigs...))
		v.bufModules = unique(append(v.bufModules, bk.grpcBufModules...))
		if len(hostRules) > 0 {
			v.hostRules = hostRules
			if err := v.Renew(); err != nil {
				return nil, err
			}
		}
		if v.operatorID == "" {
			v.operatorID = op.id
		}
		op.grpcRunners[k] = v
	}
	for k, v := range bk.cdpRunners {
		if len(hostRules) > 0 {
			v.opts = append(v.opts, hostRules.chromedpOpt())
		}
		if err := v.Renew(); err != nil {
			return nil, err
		}
		if v.operatorID == "" {
			v.operatorID = op.id
		}
		op.cdpRunners[k] = v
	}
	for k, v := range bk.sshRunners {
		if len(hostRules) > 0 {
			v.hostRules = hostRules
			if err := v.Renew(); err != nil {
				return nil, err
			}
		}
		if v.operatorID == "" {
			v.operatorID = op.id
		}
		op.sshRunners[k] = v
	}
	for k, v := range bk.includeRunners {
		op.includeRunners[k] = v
	}

	keys := map[string]struct{}{}
	for k := range op.httpRunners {
		keys[k] = struct{}{}
	}
	for k := range op.dbRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", op.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	for k := range op.grpcRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", op.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	for k := range op.cdpRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", op.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	for k := range op.sshRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", op.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	for k := range op.includeRunners {
		if _, ok := keys[k]; ok {
			return nil, fmt.Errorf("duplicate runner names (%s): %s", op.bookPath, k)
		}
		keys[k] = struct{}{}
	}
	var errs error
	for k, err := range bk.runnerErrs {
		errs = errors.Join(errs, fmt.Errorf("runner %s error: %w", k, err))
	}
	if errs != nil && !op.newOnly {
		return nil, fmt.Errorf("failed to add runners (%s): %w", op.bookPath, errs)
	}

	op.numberOfSteps = len(bk.rawSteps)

	for i, s := range bk.rawSteps {
		key := fmt.Sprintf("%d", i)
		if op.useMap {
			key = bk.stepKeys[i]
		}
		if err := op.appendStep(i, key, s); err != nil {
			if op.newOnly {
				continue
			}
			return nil, fmt.Errorf("failed to append step (%s): %w", op.bookPath, err)
		}
	}

	return op, nil
}

// appendStep appends step.
func (op *operator) appendStep(idx int, key string, s map[string]any) error {
	if op.t != nil {
		op.t.Helper()
	}
	st := newStep(idx, key, op, s)
	// if section
	if v, ok := s[ifSectionKey]; ok {
		st.ifCond, ok = v.(string)
		if !ok {
			return fmt.Errorf("invalid if condition: %v", v)
		}
		delete(s, ifSectionKey)
	}
	// desc section
	if v, ok := s[descSectionKey]; ok {
		st.desc, ok = v.(string)
		if !ok {
			return fmt.Errorf("invalid desc: %v", v)
		}
		delete(s, descSectionKey)
	}
	// defer section
	if v, ok := s[deferSectionKey]; ok {
		st.deferred, ok = v.(bool)
		if !ok {
			return fmt.Errorf("invalid defer: %v", v)
		}
		delete(s, deferSectionKey)
	}
	// force section
	if v, ok := s[forceSectionKey]; ok {
		st.force, ok = v.(bool)
		if !ok {
			return fmt.Errorf("invalid force: %v", v)
		}
		delete(s, forceSectionKey)
	}
	// loop section
	if v, ok := s[loopSectionKey]; ok {
		r, err := newLoop(v)
		if err != nil {
			return fmt.Errorf("invalid loop: %w\n%v", err, v)
		}
		st.loop = r
		delete(s, loopSectionKey)
	}
	// test runner
	if v, ok := s[testRunnerKey]; ok {
		st.testRunner = newTestRunner()
		switch vv := v.(type) {
		case bool:
			if vv {
				st.testCond = "true"
			} else {
				st.testCond = "false"
			}
		case string:
			st.testCond = vv
		default:
			return fmt.Errorf("invalid test condition: %v", v)
		}
		delete(s, testRunnerKey)
	}
	// dump runner
	if v, ok := s[dumpRunnerKey]; ok {
		st.dumpRunner = newDumpRunner()
		switch vv := v.(type) {
		case string:
			st.dumpRequest = &dumpRequest{
				expr: vv,
			}
		case map[string]any:
			expr, ok := vv["expr"]
			if !ok {
				return fmt.Errorf("invalid dump request: %v", vv)
			}
			out, ok := vv["out"]
			if !ok {
				out = "" // default: op.stdout
			}
			disableNL, ok := vv["disableTrailingNewline"]
			if !ok {
				disableNL = false
			}
			disableMask, ok := vv["disableMaskingSecrets"]
			if !ok {
				disableMask = false
			}
			st.dumpRequest = &dumpRequest{
				expr:                   cast.ToString(expr),
				out:                    cast.ToString(out),
				disableTrailingNewline: cast.ToBool(disableNL),
				disableMaskingSecrets:  cast.ToBool(disableMask),
			}
		default:
			return fmt.Errorf("invalid dump request: %v", vv)
		}
		delete(s, dumpRunnerKey)
	}
	// bind runner
	if v, ok := s[bindRunnerKey]; ok {
		st.bindRunner = newBindRunner()
		cond, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid bind condition: %v", v)
		}
		st.bindCond = cond
		delete(s, bindRunnerKey)
	}

	k, v, ok := pop(s)
	if ok {
		st.runnerKey = k
		switch {
		case k == includeRunnerKey:
			ir, err := newIncludeRunner()
			if err != nil {
				return err
			}
			st.includeRunner = ir
			c, err := parseIncludeConfig(v)
			if err != nil {
				return err
			}
			c.step = st
			st.includeConfig = c
		case k == execRunnerKey:
			st.execRunner = newExecRunner()
			vv, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid exec command: %v", v)
			}
			st.execCommand = vv
		case k == runnerRunnerKey:
			st.runnerRunner = newRunnerRunner()
			vv, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid runner runner: %v", v)
			}
			st.runnerDefinition = vv
			op.hasRunnerRunner = true
		default:
			detected := false
			h, ok := op.httpRunners[k]
			if ok {
				st.httpRunner = h
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid http request: %v", v)
				}
				st.httpRequest = vv
				detected = true
			}
			db, ok := op.dbRunners[k]
			if ok && !detected {
				st.dbRunner = db
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid db query: %v", v)
				}
				st.dbQuery = vv
				detected = true
			}
			gc, ok := op.grpcRunners[k]
			if ok && !detected {
				st.grpcRunner = gc
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid gRPC request: %v", v)
				}
				st.grpcRequest = vv
				detected = true
			}
			cc, ok := op.cdpRunners[k]
			if ok && !detected {
				st.cdpRunner = cc
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid CDP actions: %v", v)
				}
				st.cdpActions = vv
				detected = true
			}
			sc, ok := op.sshRunners[k]
			if ok && !detected {
				st.sshRunner = sc
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid SSH command: %v", v)
				}
				st.sshCommand = vv
				detected = true
			}
			ic, ok := op.includeRunners[k]
			if ok && !detected {
				st.includeRunner = ic
				c := &includeConfig{
					step: st,
				}
				st.includeConfig = c
				detected = true
			}

			if !detected {
				if !op.hasRunnerRunner {
					return fmt.Errorf("cannot find client: %s", k)
				}
				vv, ok := v.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid runner values: %v", v)
				}
				st.runnerValues = vv
			}
		}
	}

	op.steps = append(op.steps, st)
	return nil
}

// Run runbook.
func (op *operator) Run(ctx context.Context) (err error) {
	defer deprecation.PrintWarnings()
	cctx, cancel := donegroup.WithCancel(ctx)
	defer func() {
		cancel()
		var errr error
		if op.waitTimeout > 0 {
			errr = donegroup.WaitWithTimeout(cctx, op.waitTimeout)
		} else {
			errr = donegroup.Wait(cctx)
		}
		err = errors.Join(err, errr)
		op.nm.Close()
	}()
	if op.t != nil {
		op.t.Helper()
	}
	if !op.profile {
		op.sw.Disable()
	}
	opn := op.toOperatorN()
	result, err := opn.runN(cctx)
	opn.mu.Lock()
	opn.results = append(opn.results, result)
	opn.mu.Unlock()
	if err != nil {
		if !errors.Is(err, ErrFailFast) {
			return err
		}
	}
	return result.RunResults[len(result.RunResults)-1].Err
}

// DumpProfile write run time profile.
func (op *operator) DumpProfile(w io.Writer) error {
	r := op.sw.Result()
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
func (op *operator) Result() *RunResult {
	op.runResult.ID = op.runbookID()
	r := op.sw.Result()
	if r != nil {
		if err := setElasped(op.runResult, r); err != nil {
			panic(err)
		}
	}
	return op.runResult
}

func (op *operator) clearResult() {
	op.runResult = newRunResult(op.desc, op.labels, op.bookPathOrID(), op.included, op.store)
	op.runResult.ID = op.runbookID()
	for _, s := range op.steps {
		s.clearResult()
	}
}

// run - Minimum unit to run one runbook.
func (op *operator) run(ctx context.Context) error {
	defer op.sw.Start(op.trails().toProfileIDs()...).Stop()
	defer func() {
		// Results for `needs:` are not overwritten.
		_ = op.nm.TrySet(op.bookPathOrID(), op.runResult.store)
	}()
	if op.newOnly {
		return errors.New("this runbook is not allowed to run")
	}
	for k, n := range op.needs {
		select {
		case <-ctx.Done():
		case v := <-op.nm.Chan(n.path):
			op.store.needsVars[k] = v.bindVars
		}
	}
	var err error
	if op.t != nil {
		// As test helper
		op.t.Helper()
		op.t.Run(op.testName(), func(t *testing.T) {
			t.Helper()
			op.thisT = t
			if op.loop != nil {
				err = op.runLoop(ctx)
			} else {
				err = op.runInternal(ctx)
			}
			if err != nil {
				// Skip parent runner t.Error if there is an error in the included runbook
				if !errors.Is(&includedRunErr{}, err) {
					paths, indexes, errs := failedRunbookPathsAndErrors(op.runResult)
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
		op.thisT = op.t
		if err != nil {
			return fmt.Errorf("failed to run %s: %w", op.bookPathOrID(), err)
		}
		return nil
	}
	if op.loop != nil {
		err = op.runLoop(ctx)
	} else {
		err = op.runInternal(ctx)
	}
	if err != nil {
		return fmt.Errorf("failed to run %s: %w", op.bookPathOrID(), err)
	}
	return nil
}

func (op *operator) runLoop(ctx context.Context) error {
	if op.loop == nil {
		panic("invalid usage")
	}
	defer op.loop.Clear()
	retrySuccess := false
	if op.loop.Until == "" {
		retrySuccess = true
	}
	var (
		err     error
		outcome result
		bt      string
		j       int
	)
	c, err := EvalCount(op.loop.Count, op.store.toMap())
	if err != nil {
		return err
	}
	var looperr error
	for op.loop.Loop(ctx) {
		if j >= c {
			break
		}
		if j > 0 {
			// Renew runners
			for _, r := range op.cdpRunners {
				if err := r.Renew(); err != nil {
					return err
				}
			}
		}
		i := j
		op.loopIndex = &i
		trs := op.trails()
		op.capturers.setCurrentTrails(trs)
		sw := op.sw.Start(trs.toProfileIDs()...)
		err = op.runInternal(ctx)
		if err != nil {
			sw.Stop()
			looperr = errors.Join(looperr, fmt.Errorf("loop[%d]: %w", j, err))
			outcome = resultFailure
		} else {
			sw.Stop()
			if op.Skipped() {
				outcome = resultSkipped
			} else {
				outcome = resultSuccess
			}
		}
		if op.loop.Until != "" {
			store := op.store.toMap()
			store[storeStepKeyOutcome] = string(outcome)
			tf, err := EvalWithTrace(op.loop.Until, store)
			if err != nil {
				return fmt.Errorf("loop failed on %s: %w", op.bookPathOrID(), err)
			}
			if tf.OutputAsBool() {
				retrySuccess = true
				break
			} else {
				bt, err = tf.FormatTraceTree()
				if err != nil {
					return fmt.Errorf("loop failed on %s: %w", op.bookPathOrID(), err)
				}
			}
		}
		j++
	}
	if !retrySuccess {
		err := fmt.Errorf("(%s) is not true\n%s", op.loop.Until, bt)
		if op.loop.interval != nil {
			return fmt.Errorf("retry loop failed on %s.loop (count: %d, interval: %v): %w", op.bookPathOrID(), c, *op.loop.interval, err)
		} else {
			return fmt.Errorf("retry loop failed on %s.loop (count: %d, minInterval: %v, maxInterval: %v): %w", op.bookPathOrID(), c, *op.loop.minInterval, *op.loop.maxInterval, err)
		}
	}
	if op.loop.Until == "" && looperr != nil {
		// simple count
		return fmt.Errorf("loop failed on %s: %w", op.bookPathOrID(), looperr)
	}

	return nil
}

func (op *operator) runInternal(ctx context.Context) (rerr error) {
	ctx, cancel := donegroup.WithCancel(ctx)
	defer func() {
		cancel()
		rerr = errors.Join(rerr, donegroup.Wait(ctx))
	}()

	op.mu.Lock()
	defer op.mu.Unlock()
	if op.t != nil {
		op.t.Helper()
	}

	// Clear results for each scenario run (runInternal); results per root loop are not retrievable.
	op.clearResult()
	op.store.clearSteps()

	defer func() {
		// Set run error and skipped status
		op.runResult.Err = rerr
		op.runResult.Skipped = op.Skipped()
		op.runResult.StepResults = op.StepResults()

		if op.Skipped() {
			// If the scenario is skipped, beforeFuncs/afterFuncs are not executed
			return
		}

		// afterFuncs
		for i, fn := range op.afterFuncs {
			i := i
			trs := append(op.trails(), Trail{
				Type:      TrailTypeAfterFunc,
				FuncIndex: &i,
			})
			trsi := trs.toProfileIDs()
			op.sw.Start(trsi...)
			if aferr := fn(op.runResult); aferr != nil {
				rerr = newAfterFuncError(aferr)
				op.runResult.Err = rerr
			}
			op.sw.Stop(trsi...)
		}
	}()

	// context done
	select {
	case <-ctx.Done():
		if err := op.skip(); err != nil {
			rerr = err
			return
		}
		return nil
	default:
	}

	// if
	if op.ifCond != "" {
		tf, err := op.expandCondBeforeRecord(op.ifCond, &step{})
		if err != nil {
			rerr = err
			return
		}
		if !tf {
			if err := op.skip(); err != nil {
				rerr = err
				return
			}
			return nil
		}
	}

	// beforeFuncs
	for i, fn := range op.beforeFuncs {
		i := i
		trs := append(op.trails(), Trail{
			Type:      TrailTypeBeforeFunc,
			FuncIndex: &i,
		})
		trsi := trs.toProfileIDs()
		op.sw.Start(trsi...)
		if err := fn(op.runResult); err != nil {
			op.sw.Stop(trsi...)
			return newBeforeFuncError(err)
		}
		op.sw.Stop(trsi...)
	}

	// steps
	failed := false
	force := op.force
	var deferred []*deferredOpAndStep

	for _, s := range op.steps {
		if s.deferred {
			d := &deferredOpAndStep{op: op, step: s}
			deferred = append([]*deferredOpAndStep{d}, deferred...)
			op.deferred.steps = append([]*deferredOpAndStep{d}, op.deferred.steps...)
			op.record(s.idx, nil)
			continue
		}
		if failed && !force && !s.force {
			s.setResult(errStepSkipped)
			op.recordNotRun(s.idx)
			if err := op.recordResult(s.idx, resultSkipped); err != nil {
				return err
			}
			continue
		}
		err := op.runStep(ctx, s)
		s.setResult(err)
		switch {
		case errors.Is(errStepSkipped, err):
			op.recordNotRun(s.idx)
			if err := op.recordResult(s.idx, resultSkipped); err != nil {
				return err
			}
		case err != nil:
			op.recordNotRun(s.idx)
			if err := op.recordResult(s.idx, resultFailure); err != nil {
				return err
			}
			rerr = errors.Join(rerr, err)
			failed = true
		default:
			if err := op.recordResult(s.idx, resultSuccess); err != nil {
				return err
			}
		}
	}

	// deferred steps
	if op.included {
		return
	}

	for _, os := range op.deferred.steps {
		err := os.op.runStep(ctx, os.step)
		os.step.setResult(err)
		switch {
		case err != nil:
			os.op.recordNotRun(os.step.idx)
			if err := os.op.recordResult(os.step.idx, resultFailure); err != nil {
				return err
			}
			rerr = errors.Join(rerr, err)
		default:
			if err := os.op.recordResult(os.step.idx, resultSuccess); err != nil {
				return err
			}
		}
	}

	return
}

func (op *operator) bookPathOrID() string {
	if op.bookPath != "" {
		return op.bookPath
	}
	return op.id
}

func (op *operator) testName() string {
	if op.bookPath == "" {
		return fmt.Sprintf("-(%s)", op.id)
	}
	return fmt.Sprintf("%s(%s)", op.bookPath, op.id)
}

func (op *operator) stepName(i int) string {
	var prefix string

	if op.store.loopIndex != nil {
		prefix = fmt.Sprintf(".loop[%d]", *op.store.loopIndex)
	}
	if op.useMap {
		return fmt.Sprintf("%q.steps.%s%s", op.desc, op.steps[i].key, prefix)
	}

	return fmt.Sprintf("%q.steps[%d]%s", op.desc, i, prefix)
}

// expandBeforeRecord - expand before the runner records the result.
func (op *operator) expandBeforeRecord(in any, s *step) (any, error) {
	store := op.store.toMap()
	store[storeRootKeyIncluded] = op.included
	if !s.deferred {
		store[storeRootKeyPrevious] = op.store.latest()
	}
	return EvalExpand(in, store)
}

// expandCondBeforeRecord - expand condition before the runner records the result.
func (op *operator) expandCondBeforeRecord(ifCond string, s *step) (bool, error) {
	store := op.store.toMap()
	store[storeRootKeyIncluded] = op.included
	if !s.deferred {
		store[storeRootKeyPrevious] = op.store.latest()
	}
	return EvalCond(ifCond, store)
}

// Debugln print to out when debug = true.
func (op *operator) Debugln(a any) {
	if !op.debug {
		return
	}
	_, _ = fmt.Fprintln(op.stderr, a)
}

// Debugf print to out when debug = true.
func (op *operator) Debugf(format string, a ...any) {
	if !op.debug {
		return
	}
	_, _ = fmt.Fprintf(op.stderr, format, a...)
}

// Warnf print to out.
func (op *operator) Warnf(format string, a ...any) {
	_, _ = fmt.Fprintf(op.stderr, format, a...)
}

// Skipped returns whether the runbook run skipped.
func (op *operator) Skipped() bool {
	return op.skipped
}

func (op *operator) skip() error {
	op.Debugf(yellow("Skip %s\n"), op.desc)
	op.skipped = true
	for i, s := range op.steps {
		s.setResult(errStepSkipped)
		op.recordNotRun(i)
		if err := op.recordResult(s.idx, resultSkipped); err != nil {
			return err
		}
	}
	return nil
}

// toOperatorN convert *operator top *operatorN.
func (op *operator) toOperatorN() *operatorN {
	opn := &operatorN{
		ops:       []*operator{op},
		om:        map[string]*operator{},
		nm:        op.nm,
		included:  map[string][]string{},
		t:         op.t,
		sw:        op.sw,
		profile:   op.profile,
		concmax:   1,
		kv:        op.store.kv,
		runNIndex: atomic.Int64{},
		opts:      op.exportOptionsToBePropagated(),
		dbg:       op.dbg,
	}
	opn.runNIndex.Store(-1)
	opn.dbg.setOperatorN(opn) // link back to dbg

	_ = opn.traverseOperators(op)

	return opn
}

func (op *operator) StepResults() []*StepResult {
	var results []*StepResult
	for _, s := range op.steps {
		if lo.ContainsBy(op.deferred.steps, func(op *deferredOpAndStep) bool {
			return s.runbookID() == op.step.runbookID()
		}) {
			continue
		}
		results = append(results, s.result)
	}
	for _, os := range op.deferred.steps {
		if op.id == os.op.id {
			results = append(results, os.step.result)
		}
	}
	return results
}

type operatorN struct {
	ops          []*operator                      // All operators without `needs:` that may run.
	om           map[string]*operator             // Map of all operatorN traversed including `needs:`. Use like cache
	nm           *waitmap.WaitMap[string, *store] // Map of runbook result stores. key is the operator.bookPath.
	skipIncluded bool                             // Skip running the included runbook by itself.
	included     map[string][]string              // Runbook paths included by another runbooks. map[includedRunbookPath] = []string{includingRunbookPath}.
	t            *testing.T
	sw           *stopw.Span
	profile      bool          // profile is the flag to enable profiling.
	shuffle      bool          // shuffle is the flag to shuffle the operators.
	shuffleSeed  int64         // shuffleSeed is the seed for shuffling the operators.
	shardN       int           // shardN is the number of shards to run.
	shardIndex   int           // shardIndex is the index of the shard to run.
	sample       int           // sample is the number of operators to run.
	random       int           // random is the number of operators to run randomly.
	waitTimeout  time.Duration // waitTimout is the time to wait for sub-processes to complete after the Run or RunN context is canceled.
	concmax      int
	failFast     bool
	opts         []Option
	results      []*runNResult
	runNIndex    atomic.Int64 // runNIndex holds the runN execution index (starting from 0). It is incremented each time runN is executed
	kv           *kv.KV
	dbg          *dbg
	mu           sync.Mutex
}

func Load(pathp string, opts ...Option) (*operatorN, error) {
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
	opn := &operatorN{
		om:           map[string]*operator{},
		nm:           waitmap.New[string, *store](),
		skipIncluded: bk.skipIncluded,
		included:     map[string][]string{},
		t:            bk.t,
		sw:           sw,
		profile:      bk.profile,
		shuffle:      bk.runShuffle,
		shuffleSeed:  bk.runShuffleSeed,
		shardN:       bk.runShardN,
		shardIndex:   bk.runShardIndex,
		sample:       bk.runSample,
		random:       bk.runRandom,
		waitTimeout:  bk.waitTimeout,
		failFast:     bk.failFast,
		concmax:      1,
		opts:         opts,
		runNIndex:    atomic.Int64{},
		kv:           kv.New(),
		dbg:          newDBG(bk.attach),
	}
	opn.runNIndex.Store(-1) // Set index to -1 ( no runN )

	opn.dbg.setOperatorN(opn) // link back to dbg
	if bk.runConcurrent {
		opn.concmax = bk.runConcurrentMax
	}
	books, err := Books(pathp)
	if err != nil {
		return nil, err
	}
	var loaded []*operator // loaded operatorN without `needs:` that may run.
	for _, b := range books {
		o, err := New(append([]Option{b}, opts...)...)
		if err != nil {
			return nil, err
		}
		if err := opn.traverseOperators(o); err != nil {
			return nil, err
		}
		loaded = append(loaded, o)
	}

	// Generate IDs for all operatorN that may run.
	if err := opn.generateIDsUsingPath(); err != nil {
		return nil, err
	}

	var idMatched []*operator
	cond := labelCond(bk.runLabels)
	indexes := map[string]int{}
	opn.ops = nil
	for _, op := range loaded {
		p := op.bookPath
		// RUNN_RUN, --run
		if !bk.runMatch.MatchString(p) {
			op.Debugf(yellow("Skip %s because it does not match %s\n"), p, bk.runMatch.String())
			continue
		}
		// RUNN_LABEL, --label
		tf, err := EvalCond(cond, labelEnv(op.labels))
		if err != nil {
			return nil, err
		}
		if !tf {
			op.Debugf(yellow("Skip %s because it does not match %s\n"), p, cond)
			continue
		}
		// RUNN_ID, --id
		for i, id := range bk.runIDs {
			if strings.HasPrefix(op.id, id) {
				idMatched = append(idMatched, op)
				indexes[op.id] = i
			}
		}
		op.sw = opn.sw
		op.nm = opn.nm
		opn.ops = append(opn.ops, op)
	}

	// Run the matching runbooks in order if there is only one runbook with a forward matching ID.
	if len(bk.runIDs) > 0 {
		switch {
		case len(idMatched) == 0:
			return nil, fmt.Errorf("no runbooks has the id prefix: %s", bk.runIDs)
		default:
			u := lo.UniqBy(idMatched, func(op *operator) string {
				return op.id
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
			opn.ops = idMatched
		}
	} else {
		// If no ids are specified, the order is sorted and fixed
		sortOperators(opn.ops)
	}
	if err := opn.skipIncludedOperators(); err != nil {
		return nil, err
	}
	return opn, nil
}

func (opn *operatorN) RunN(ctx context.Context) (err error) {
	defer deprecation.PrintWarnings()
	cctx, cancel := donegroup.WithCancel(ctx)
	defer func() {
		cancel()
		var errr error
		if opn.waitTimeout > 0 {
			errr = donegroup.WaitWithTimeout(cctx, opn.waitTimeout)
		} else {
			errr = donegroup.Wait(cctx)
		}
		err = errors.Join(err, errr)
		opn.nm.Close()
	}()
	if opn.t != nil {
		opn.t.Helper()
	}
	if !opn.profile {
		opn.sw.Disable()
	}
	result, err := opn.runN(cctx)
	opn.mu.Lock()
	opn.results = append(opn.results, result)
	opn.mu.Unlock()
	if err != nil {
		if !errors.Is(err, ErrFailFast) {
			return err
		}
	}
	return nil
}

func (opn *operatorN) Operators() []*operator {
	return opn.ops
}

func (opn *operatorN) Close() {
	for _, op := range opn.ops {
		op.Close(true)
	}
}

func (opn *operatorN) DumpProfile(w io.Writer) error {
	r := opn.sw.Result()
	if r == nil {
		return errors.New("no profile")
	}
	enc := ejson.NewEncoder(w)
	if err := enc.Encode(r); err != nil {
		return err
	}
	return nil
}

func (opn *operatorN) Init() error {
	return nil
}

func (opn *operatorN) RequestOne(ctx context.Context) error {
	if !opn.profile {
		opn.sw.Disable()
	}
	ctx = context.WithoutCancel(ctx)
	result, err := opn.runN(ctx)
	if err != nil {
		return err
	}
	if result.HasFailure() {
		return errors.New("result has failure")
	}
	return nil
}

func (opn *operatorN) Terminate() error {
	opn.Close()
	return nil
}

func (opn *operatorN) Result() *runNResult {
	return opn.results[len(opn.results)-1]
}

func (opn *operatorN) SelectedOperators() (tops []*operator, err error) {
	defer func() {
		selected := &operatorN{
			ops:          tops,
			sw:           opn.sw,
			om:           opn.om,
			nm:           opn.nm,
			skipIncluded: opn.skipIncluded,
			included:     map[string][]string{},
			t:            opn.t,
			opts:         opn.opts,
			kv:           opn.kv,
			dbg:          opn.dbg,
		}
		for _, op := range tops {
			if errr := selected.traverseOperators(op); errr != nil {
				err = errors.Join(err, errr)
			}
		}
		if err == nil {
			tops, err = sortWithNeeds(selected.ops)
		}
	}()

	tops = make([]*operator, len(opn.ops))
	copy(tops, opn.ops)
	if opn.runNIndex.Load() > 0 && opn.random == 0 {
		// Copy operators for each runN
		tops, err = copyOperators(tops, opn.opts)
		if err != nil {
			return nil, err
		}
	}
	if opn.shuffle {
		// Shuffle order of running
		shuffleOperators(tops, opn.shuffleSeed)
	}

	if opn.shardN > 0 {
		tops = partOperators(tops, opn.shardN, opn.shardIndex)
	}
	if opn.sample > 0 {
		tops = sampleOperators(tops, opn.sample)
	}
	if opn.random > 0 {
		rops, err := randomOperators(tops, opn.opts, opn.random)
		if err != nil {
			return nil, err
		}
		for _, op := range rops {
			op.sw = opn.sw
		}
		return rops, nil
	}

	return tops, nil
}

func (opn *operatorN) CollectCoverage(ctx context.Context) (*Coverage, error) {
	cov := &Coverage{}
	for _, op := range opn.ops {
		c, err := op.collectCoverage(ctx)
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

// SetKV sets a key-value pair to runn.kv.
func (opn *operatorN) SetKV(k string, v any) {
	opn.kv.Set(k, v)
}

// GetKV gets a value from runn.kv.
func (opn *operatorN) GetKV(k string) any { //nostyle:getters
	return opn.kv.Get(k)
}

// DelKV deletes a key-value pair from runn.kv.
func (opn *operatorN) DelKV(k string) {
	opn.kv.Del(k)
}

// ClearKV clears all key-value pairs in runn.kv.
func (opn *operatorN) Clear() {
	opn.kv.Clear()
}

func (opn *operatorN) runN(ctx context.Context) (*runNResult, error) {
	result := &runNResult{}
	if opn.t != nil {
		opn.t.Helper()
	}
	defer opn.sw.Start().Stop()
	defer opn.Close()
	runNIndex := opn.runNIndex.Add(1)
	cg, cctx := concgroup.WithContext(ctx)
	cg.SetLimit(opn.concmax)
	selected, err := opn.SelectedOperators()
	if err != nil {
		return result, err
	}
	result.Total.Add(int64(len(selected)))
	for _, op := range selected {
		op := op
		op.store.runNIndex = int(runNIndex) // Set runN index
		cg.GoMulti(op.concurrency, func() error {
			defer func() {
				r := op.Result()
				op.capturers.captureResult(op.trails(), r)
				op.capturers.captureEnd(op.trails(), op.bookPath, op.desc)
				op.Close(false)
				result.mu.Lock()
				result.RunResults = append(result.RunResults, r)
				result.mu.Unlock()
			}()
			op.capturers.captureStart(op.trails(), op.bookPath, op.desc)
			if err := op.run(cctx); err != nil {
				if opn.failFast {
					return errors.Join(err, ErrFailFast)
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

// traverseOperators traverse operator(s) recursively.
func (opn *operatorN) traverseOperators(op *operator) error {
	defer func() {
		opn.ops = lo.UniqBy(opn.ops, func(op *operator) string {
			return op.bookPathOrID()
		})
	}()

	for _, oo := range opn.ops {
		if _, ok := opn.om[oo.bookPath]; !ok {
			opn.om[oo.bookPath] = oo
		}
	}

	// needs:
	paths := lo.MapToSlice(op.needs, func(_ string, n *need) string {
		return n.path
	})

	for _, p := range paths {
		if oo, ok := opn.om[p]; ok {
			// already loaded
			opn.ops = append([]*operator{oo}, opn.ops...)
			for k, n := range op.needs {
				if n.path == p && op.needs[k].op == nil {
					op.needs[k].op = oo
				}
			}
			continue
		}
		needo, err := New(append([]Option{Book(p)}, opn.opts...)...)
		if err != nil {
			return err
		}
		opn.om[p] = needo
		needo.store.kv = opn.kv // set pointer of kv
		needo.dbg = opn.dbg

		for k, n := range op.needs {
			if n.path == p && op.needs[k].op == nil {
				op.needs[k].op = needo
			}
		}

		if err := opn.traverseOperators(needo); err != nil {
			return err
		}
		opn.ops = append([]*operator{needo}, opn.ops...)
	}

	for _, s := range op.steps {
		if s.includeRunner != nil && s.includeConfig != nil {
			p, err := fp(s.includeConfig.path, op.root)
			if err != nil {
				return err
			}
			if _, ok := opn.included[p]; !ok {
				opn.included[p] = []string{}
			}
			opn.included[p] = append(opn.included[p], op.bookPath)
		}
	}

	op.store.kv = opn.kv // set pointer of kv
	op.dbg = opn.dbg
	op.nm = opn.nm
	op.sw = opn.sw

	if _, ok := opn.om[op.bookPath]; !ok {
		opn.om[op.bookPath] = op
	}

	return nil
}

// skipIncludedOperators skips operators that are included by other operators.
func (opn *operatorN) skipIncludedOperators() error {
	if !opn.skipIncluded {
		return nil
	}
	if len(opn.included) == 0 {
		return nil
	}
	var filtered []*operator
	var filteredPaths []string
	for _, op := range opn.ops {
		if _, ok := opn.included[op.bookPath]; ok {
			continue
		}
		filtered = append(filtered, op)
		filteredPaths = append(filteredPaths, op.bookPath)
	}
L:
	for _, op := range opn.ops {
		including, ok := opn.included[op.bookPath]
		if !ok {
			continue
		}
		for _, inc := range including {
			if slices.Contains(filteredPaths, inc) {
				continue L
			}
		}
		filtered = append(filtered, op)
	}

	opn.ops = filtered
	return nil
}

// sortWithNeeds sort operatorN after resolving dependencies by `needs:`.
func sortWithNeeds(ops []*operator) ([]*operator, error) {
	var sorted []*operator
	for _, op := range ops {
		needs, err := resolveNeeds(op, 0)
		if err != nil {
			return nil, err
		}
		sorted = append(sorted, needs...)
	}
	return lo.Uniq(sorted), nil
}

func resolveNeeds(op *operator, depth int) ([]*operator, error) {
	const maxDepth = 10
	if depth > maxDepth {
		return nil, fmt.Errorf("`needs:` max depth exceeded: %d", maxDepth)
	}
	if len(op.needs) == 0 {
		return []*operator{op}, nil
	}
	var needs []*operator
	for _, n := range op.needs {
		resolved, err := resolveNeeds(n.op, depth+1)
		if err != nil {
			return nil, err
		}
		needs = append(resolved, needs...)
	}
	needs = append(needs, op)
	return needs, nil
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
	for _, op := range ops {
		// FIXME: Need the function to copy the operator as it is heavy to parse the runbook each time
		oo, err := New(append([]Option{Book(op.bookPath)}, opts...)...)
		if err != nil {
			return nil, err
		}
		oo.id = op.id // Copy id from original operator
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
		op, err := New(append([]Option{Book(n[idx].bookPath)}, opts...)...)
		if err != nil {
			return nil, err
		}
		op.id = ops[idx].id // Copy id from original operator
		random = append(random, op)
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
		for _, ir := range sr.IncludedRunResults {
			if err := setElaspedByRunbookIDFull(ir, m); err != nil {
				return err
			}
		}
	}
	return nil
}

var labelRep = strings.NewReplacer("-", "___hyphen___", "/", "___slash___", ".", "___dot___", ":", "___colon___")

func labelEnv(labels []string) exprtrace.EvalEnv {
	labelsMap := lo.SliceToMap(labels, func(l string) (string, bool) {
		return labelRep.Replace(l), true
	})
	return exprtrace.EvalEnv{
		"labels": labelsMap,
	}
}

func labelCond(labels []string) string {
	if len(labels) == 0 {
		return "true"
	}
	var sb strings.Builder
	for i, label := range labels {
		if i > 0 {
			sb.WriteString(" or ")
		}

		label = strings.ReplaceAll(label, "!", "not ")

		sb.WriteString("(")
		for _, s := range strings.Split(label, " ") {
			switch s {
			case "not":
				sb.WriteString("not ")
			case "or":
				sb.WriteString(" or ")
			case "and":
				sb.WriteString(" and ")
			default:
				sb.WriteString("labels.")
				sb.WriteString(labelRep.Replace(s))
			}
		}
		sb.WriteString(")")
	}

	return sb.String()
}
