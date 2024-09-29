package runn

import (
	"context"
	"errors"
	"path/filepath"
)

const includeRunnerKey = "include"

type includeRunner struct {
	name       string
	path       string
	params     map[string]any
	runResults []*RunResult
}

type includeConfig struct {
	path     string
	vars     map[string]any
	skipTest bool
	force    bool
	step     *step
}

type includedRunErr struct {
	err error
}

func newIncludedRunErr(err error) *includedRunErr {
	return &includedRunErr{err: err}
}

func (e *includedRunErr) Error() string {
	return e.err.Error()
}

func (e *includedRunErr) Unwrap() error {
	return e.err
}

func (e *includedRunErr) Is(target error) bool {
	err := target
	for {
		_, ok := err.(*includedRunErr) //nolint:errorlint
		if ok {
			return true
		}
		if err = errors.Unwrap(err); err == nil {
			break
		}
	}
	return false
}

func newIncludeRunner() (*includeRunner, error) {
	return &includeRunner{}, nil
}

func (rnr *includeRunner) Run(ctx context.Context, s *step) error {
	o := s.parent
	c := s.includeConfig
	if o.thisT != nil {
		o.thisT.Helper()
	}
	rnr.runResults = nil

	ipath := rnr.path
	if ipath == "" {
		ipath = c.path
	}
	// ipath must not be variable expanded. Because it will be impossible to identify the step of the included runbook in case of run failure.
	if !hasRemotePrefix(ipath) {
		ipath = filepath.Join(o.root, ipath)
	}

	// Store before record
	store := o.store.toMap()
	store[storeRootKeyIncluded] = o.included
	store[storeRootKeyPrevious] = o.store.latest()

	nodes, err := s.expandNodes()
	if err != nil {
		return err
	}
	if rnr.name != "" {
		v, ok := nodes[rnr.name].(map[string]any)
		if ok {
			store[storeRootKeyNodes] = v
		}
	}

	params := map[string]any{}
	for k, v := range rnr.params {
		switch ov := v.(type) {
		case string:
			var vv any
			vv, err = o.expandBeforeRecord(ov)
			if err != nil {
				return err
			}
			evv, err := evaluateSchema(vv, o.root, store)
			if err != nil {
				return err
			}
			params[k] = evv
		case map[string]any, []any:
			vv, err := o.expandBeforeRecord(ov)
			if err != nil {
				return err
			}
			params[k] = vv
		default:
			params[k] = ov
		}
	}
	if len(params) > 0 {
		store[storeRootKeyParams] = params
	}

	pstore := map[string]any{
		storeRootKeyParent: store,
	}

	oo, err := o.newNestedOperator(c.step, bookWithStore(ipath, pstore), SkipTest(c.skipTest))
	if err != nil {
		return err
	}

	// Override vars
	for k, v := range c.vars {
		switch ov := v.(type) {
		case string:
			var vv any
			vv, err = o.expandBeforeRecord(ov)
			if err != nil {
				return err
			}
			evv, err := evaluateSchema(vv, oo.root, store)
			if err != nil {
				return err
			}
			oo.store.vars[k] = evv
		case map[string]any, []any:
			vv, err := o.expandBeforeRecord(ov)
			if err != nil {
				return err
			}
			oo.store.vars[k] = vv
		default:
			oo.store.vars[k] = ov
		}
	}

	if err := rnr.run(ctx, oo, s); err != nil {
		return err
	}
	return nil
}

func (rnr *includeRunner) run(ctx context.Context, oo *operator, s *step) error {
	o := s.parent

	ops := oo.toOperatorN()
	sorted, err := sortWithNeeds(ops.ops)
	if err != nil {
		return err
	}
	// Filter already runned runbooks
	var filtered []*operator
	for _, ooo := range sorted {
		if oo.bookPath == ooo.bookPath {
			// The originally included oo is not filtered.
			filtered = append(filtered, ooo)
		} else if _, ok := ops.nm.TryGet(ooo.bookPath); !ok {
			filtered = append(filtered, ooo)
		}
	}

	// Do not use ops.runN because runN closes the runners.
	// And one runbook should be run sequentially.
	// ref: https://github.com/k1LoW/runn/blob/b81205550f0e15fec509a596fcee8619e345ae95/docs/designs/id.md
	for _, ooo := range filtered {
		ooo.parent = oo.parent
		if err := ooo.run(ctx); err != nil {
			rnr.runResults = append(rnr.runResults, ooo.runResult)
			return newIncludedRunErr(err)
		}
		rnr.runResults = append(rnr.runResults, ooo.runResult)
	}
	o.record(oo.store.toNormalizedMap())
	return nil
}

// newNestedOperator create nested operator.
func (o *operator) newNestedOperator(parent *step, opts ...Option) (*operator, error) {
	popts := append([]Option{included(true)}, o.exportOptionsToBePropagated()...)

	// Prefer child runbook opts
	// For example, if a runner with the same name is defined in the child runbook to be included, it takes precedence.
	opts = append(popts, opts...)
	oo, err := New(opts...)
	if err != nil {
		return nil, err
	}
	// Nested operatorN do not inherit beforeFuncs/afterFuncs
	oo.t = o.thisT
	oo.thisT = o.thisT
	oo.sw = o.sw
	oo.capturers = o.capturers
	oo.parent = parent
	oo.store.parentVars = o.store.toMap()
	oo.store.kv = o.store.kv
	oo.dbg = o.dbg
	oo.nm = o.nm
	return oo, nil
}

// export exports options.
func (o *operator) exportOptionsToBePropagated() []Option {
	var opts []Option

	// Set parent runners for re-use
	for k, r := range o.httpRunners {
		opts = append(opts, reuseHTTPRunner(k, r))
	}
	for k, r := range o.dbRunners {
		opts = append(opts, reuseDBRunner(k, r))
	}
	for k, r := range o.grpcRunners {
		opts = append(opts, reuseGrpcRunner(k, r))
	}
	for k, r := range o.sshRunners {
		opts = append(opts, reuseSSHRunner(k, r))
	}

	opts = append(opts, Debug(o.debug))
	opts = append(opts, Profile(o.profile))
	opts = append(opts, SkipTest(o.skipTest))
	opts = append(opts, Force(o.force))
	opts = append(opts, Trace(o.trace))
	for k, f := range o.store.funcs {
		opts = append(opts, Func(k, f))
	}
	return opts
}
