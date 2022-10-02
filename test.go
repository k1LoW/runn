package runn

import (
	"context"
	"fmt"
)

const testRunnerKey = "test"

type testRunner struct {
	operator *operator
}

func newTestRunner(o *operator) (*testRunner, error) {
	return &testRunner{
		operator: o,
	}, nil
}

func (rnr *testRunner) Run(ctx context.Context, cond string, runned bool) error {
	store := rnr.operator.store.toMap()
	if runned {
		store[storeCurrentKey] = rnr.operator.store.latest()
	}
	t, err := buildTree(cond, store)
	if err != nil {
		return err
	}
	rnr.operator.Debugln("-----START TEST CONDITION-----")
	rnr.operator.Debugf("%s", t)
	rnr.operator.Debugln("-----END TEST CONDITION-----")
	tf, err := evalCond(cond, store)
	if err != nil {
		return err
	}
	if !tf {
		return fmt.Errorf("(%s) is not true\n%s", cond, t)
	}
	return nil
}
