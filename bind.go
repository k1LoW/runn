package runn

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
)

const bindRunnerKey = "bind"

type bindRunner struct {
	operator *operator
}

func newBindRunner(o *operator) (*bindRunner, error) {
	return &bindRunner{
		operator: o,
	}, nil
}

func (rnr *bindRunner) Run(ctx context.Context, cond map[string]string) error {
	store := rnr.operator.store.toMap()
	for k, v := range cond {
		if k == storeVarsKey || k == storeStepsKey || k == storeParentKey {
			return fmt.Errorf("'%s' is reserved", k)
		}
		vv, err := expr.Eval(v, store)
		if err != nil {
			return err
		}
		rnr.operator.store.bindVars[k] = vv
	}
	return nil
}
