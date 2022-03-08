package runn

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
)

type testRunner struct {
	operator *operator
}

func newTestRunner(o *operator) (*testRunner, error) {
	return &testRunner{
		operator: o,
	}, nil
}

func (rnr *testRunner) Run(ctx context.Context, cond string) error {
	store := map[string]interface{}{
		"steps": rnr.operator.store.steps,
		"vars":  rnr.operator.store.vars,
	}
	tf, err := expr.Eval(fmt.Sprintf("(%s) == true", cond), store)
	if err != nil {
		return err
	}
	rnr.operator.store.steps = append(rnr.operator.store.steps, nil)
	if !tf.(bool) {
		return fmt.Errorf("(%s) is false", cond)
	}
	return nil
}
