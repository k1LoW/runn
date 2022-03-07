package runbk

import (
	"context"
	"fmt"

	"github.com/antonmedv/expr"
)

type testRunner struct {
	operator *Operator
}

func newTestRunner(o *Operator) (*testRunner, error) {
	return &testRunner{
		operator: o,
	}, nil
}

func (c *testRunner) Run(ctx context.Context, cond string) error {
	store := map[string]interface{}{
		"steps": c.operator.store.Steps,
		"vars":  c.operator.store.Vars,
	}
	tf, err := expr.Eval(fmt.Sprintf("(%s) == true", cond), store)
	if err != nil {
		return err
	}
	c.operator.store.Steps = append(c.operator.store.Steps, nil)
	if !tf.(bool) {
		return fmt.Errorf("(%s) is false", cond)
	}
	return nil
}
