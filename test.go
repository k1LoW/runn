package runbk

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

func (c *testRunner) Run(ctx context.Context, cond string) error {
	store := map[string]interface{}{
		"steps": c.operator.store.steps,
		"vars":  c.operator.store.vars,
	}
	tf, err := expr.Eval(fmt.Sprintf("(%s) == true", cond), store)
	if err != nil {
		return err
	}
	c.operator.store.steps = append(c.operator.store.steps, nil)
	if !tf.(bool) {
		return fmt.Errorf("(%s) is false", cond)
	}
	return nil
}
