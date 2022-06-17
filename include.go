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
	oo, err := rnr.operator.newNestedOperator(Book(filepath.Join(rnr.operator.root, c.path)), SkipTest(c.skipTest))
	if err != nil {
		return err
	}
	// override vars
	for k, v := range c.vars {
		vv, err := rnr.operator.expand(v)
		if err != nil {
			return err
		}
		oo.store.vars[k] = vv
	}
	if err := oo.Run(ctx); err != nil {
		return err
	}
	rnr.operator.record(oo.store.toMap())

	for _, r := range oo.httpRunners {
		r.operator = rnr.operator
	}
	for _, r := range oo.dbRunners {
		r.operator = rnr.operator
	}

	return nil
}

func (o *operator) newNestedOperator(opts ...Option) (*operator, error) {
	opts = append(opts, included(true))

	for k, r := range o.httpRunners {
		opts = append(opts, runnHTTPRunner(k, r))
	}
	for k, r := range o.dbRunners {
		opts = append(opts, runnDBRunner(k, r))
	}
	opts = append(opts, Var("parent", o.store.steps))
	opts = append(opts, Debug(o.debug))
	opts = append(opts, SkipTest(o.skipTest))
	oo, err := New(opts...)
	if err != nil {
		return nil, err
	}
	oo.t = o.thisT
	oo.thisT = o.thisT
	return oo, nil
}
