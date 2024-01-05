package runn

import (
	"context"
	"errors"
	"path/filepath"
)

const includeRunnerKey = "include"

type includeRunner struct {
	runResult *RunResult
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
	rnr.runResult = nil

	// c.path must not be variable expanded. Because it will be impossible to identify the step of the included runbook in case of run failure.
	var ibp string
	if hasRemotePrefix(c.path) {
		ibp = c.path
	} else {
		ibp = filepath.Join(o.root, c.path)
	}

	// Store before record
	store := o.store.toMap()
	store[storeRootKeyIncluded] = o.included
	store[storeRootPrevious] = o.store.latest()
	pstore := map[string]any{
		storeRootKeyParent: store,
	}
	oo, err := o.newNestedOperator(c.step, bookWithStore(ibp, pstore), SkipTest(c.skipTest))
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
	if err := oo.run(ctx); err != nil {
		rnr.runResult = oo.runResult
		return newIncludedRunErr(err)
	}
	rnr.runResult = oo.runResult
	o.record(oo.store.toNormalizedMap())
	return nil
}

// newNestedOperator create nested operator.
func (o *operator) newNestedOperator(parent *step, opts ...Option) (*operator, error) {
	var popts []Option
	popts = append(popts, included(true))

	// Set parent runners for re-use
	for k, r := range o.httpRunners {
		popts = append(popts, runnHTTPRunner(k, r))
	}
	for k, r := range o.dbRunners {
		popts = append(popts, runnDBRunner(k, r))
	}
	for k, r := range o.grpcRunners {
		popts = append(popts, runnGrpcRunner(k, r))
	}
	for k, r := range o.sshRunners {
		popts = append(popts, runnSSHRunner(k, r))
	}

	popts = append(popts, Debug(o.debug))
	popts = append(popts, Profile(o.profile))
	popts = append(popts, SkipTest(o.skipTest))
	popts = append(popts, Force(o.force))
	popts = append(popts, Trace(o.trace))
	for k, f := range o.store.funcs {
		popts = append(popts, Func(k, f))
	}
	// Prefer child runbook opts
	opts = append(popts, opts...)
	oo, err := New(opts...)
	if err != nil {
		return nil, err
	}
	// Nested operators do not inherit beforeFuncs/afterFuncs
	oo.t = o.thisT
	oo.thisT = o.thisT
	oo.sw = o.sw
	oo.capturers = o.capturers
	oo.parent = parent
	oo.store.parentVars = o.store.toMap()
	return oo, nil
}
