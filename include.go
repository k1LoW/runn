package runn

import (
	"context"
	"path/filepath"
)

const includeRunnerKey = "include"

type includeRunner struct {
	operator *operator
}

func newIncludeRunner(o *operator) (*includeRunner, error) {
	return &includeRunner{
		operator: o,
	}, nil
}

func (rnr *includeRunner) Run(ctx context.Context, path string) error {
	oo, err := rnr.operator.newNestedOperator(Book(filepath.Join(rnr.operator.root, path)))
	if err != nil {
		return err
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
	for k, r := range o.httpRunners {
		opts = append(opts, runnHTTPRunner(k, r))
	}
	for k, r := range o.dbRunners {
		opts = append(opts, runnDBRunner(k, r))
	}
	for k, v := range o.store.vars {
		opts = append(opts, Var(k, v))
	}
	opts = append(opts, Var("parent", o.store.steps))
	opts = append(opts, Debug(o.debug))
	oo, err := New(opts...)
	if err != nil {
		return nil, err
	}
	oo.t = o.t
	return oo, nil
}
