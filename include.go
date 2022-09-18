package runn

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/antonmedv/expr"
)

const includeRunnerKey = "include"

type includeRunner struct {
	operator *operator
}

type includeConfig struct {
	path     string
	vars     map[string]interface{}
	skipTest bool
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
	oo, err := rnr.operator.newNestedOperator(c.step, Book(filepath.Join(rnr.operator.root, c.path)), SkipTest(c.skipTest))
	if err != nil {
		return err
	}
	store := rnr.operator.store.toMap()
	// override vars
	for k, v := range c.vars {
		switch o := v.(type) {
		case string:
			var vv interface{}
			if !strings.Contains(o, delimStart) {
				vv = o
			} else {
				matches := expandRe.FindAllStringSubmatch(o, -1)
				if len(matches) > 1 || !strings.HasPrefix(o, delimStart) || !strings.HasSuffix(o, delimEnd) {
					vv, err = rnr.operator.expand(o)
					if err != nil {
						return err
					}
				} else {
					vv, err = expr.Eval(matches[0][1], store)
					if err != nil {
						return err
					}
				}
			}

			evv, err := evaluateSchema(vv, oo.root, store)
			if err != nil {
				return err
			}

			oo.store.vars[k] = evv
		case map[string]interface{}, []interface{}:
			vv, err := rnr.operator.expand(o)
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

	for _, r := range oo.httpRunners {
		r.operator = rnr.operator
	}
	for _, r := range oo.dbRunners {
		r.operator = rnr.operator
	}

	return nil
}

func (o *operator) newNestedOperator(parent *step, opts ...Option) (*operator, error) {
	opts = append(opts, included(true))
	for k, r := range o.httpRunners {
		opts = append(opts, runnHTTPRunner(k, r))
	}
	for k, r := range o.dbRunners {
		opts = append(opts, runnDBRunner(k, r))
	}
	for k, r := range o.grpcRunners {
		opts = append(opts, runnGrpcRunner(k, r))
	}
	opts = append(opts, Debug(o.debug))
	opts = append(opts, Profile(o.profile))
	opts = append(opts, SkipTest(o.skipTest))
	for k, f := range o.store.funcs {
		opts = append(opts, Func(k, f))
	}
	oo, err := New(opts...)
	if err != nil {
		return nil, err
	}
	oo.t = o.thisT
	oo.thisT = o.thisT
	oo.sw = o.sw
	oo.capturers = o.capturers
	oo.parent = parent
	oo.store.parentVars = o.store.toMap()
	return oo, nil
}
