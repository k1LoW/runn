package runn

import (
	"context"
	"path/filepath"
)

const includeRunnerKey = "include"

type includeRunner struct {
	operator *operator
}

type includeConfig struct {
	path     string
	vars     map[string]interface{}
	skipTest bool
	force    bool
	step     *step
}

func newIncludeRunner(o *operator) (*includeRunner, error) {
	return &includeRunner{
		operator: o,
	}, nil
}

func (rnr *includeRunner) Run(ctx context.Context, c *includeConfig) error {
	if rnr.operator.thisT != nil {
		rnr.operator.thisT.Helper()
	}
	ibp := filepath.Join(rnr.operator.root, c.path)
	if err := fetchFile(ibp); err != nil {
		return err
	}

	// Store before record
	store := rnr.operator.store.toMap()
	store[storeIncludedKey] = rnr.operator.included
	store[storePreviousKey] = rnr.operator.store.latest()
	pstore := map[string]interface{}{
		storeParentKey: store,
	}
	oo, err := rnr.operator.newNestedOperator(c.step, bookWithStore(ibp, pstore), SkipTest(c.skipTest))
	if err != nil {
		return err
	}

	// Override vars
	for k, v := range c.vars {
		switch o := v.(type) {
		case string:
			var vv interface{}
			vv, err = rnr.operator.expandBeforeRecord(o)
			if err != nil {
				return err
			}
			evv, err := evaluateSchema(vv, oo.root, store)
			if err != nil {
				return err
			}
			oo.store.vars[k] = evv
		case map[string]interface{}, []interface{}:
			vv, err := rnr.operator.expandBeforeRecord(o)
			if err != nil {
				return err
			}
			oo.store.vars[k] = vv
		default:
			oo.store.vars[k] = o
		}
	}
	if err := oo.run(ctx); err != nil {
		return err
	}
	rnr.operator.record(oo.store.toNormalizedMap())

	// Restore the condition of runners re-used in child runbooks.
	for _, r := range oo.httpRunners {
		r.operator = rnr.operator
	}
	for _, r := range oo.dbRunners {
		r.operator = rnr.operator
	}
	for _, r := range oo.grpcRunners {
		r.operator = rnr.operator
	}
	for _, r := range oo.sshRunners {
		r.operator = rnr.operator
	}

	return nil
}

// newNestedOperator create nested operator.
func (o *operator) newNestedOperator(parent *step, opts ...Option) (*operator, error) {
	popts := []Option{}
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
