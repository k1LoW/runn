package runn

import (
	"context"
	"fmt"
)

const testRunnerKey = "test"

type testRunner struct {
	operator *operator
}

type condFalseError struct {
	cond string
	tree string
}

func newCondFalseError(cond, tree string) *condFalseError {
	return &condFalseError{
		cond: cond,
		tree: tree,
	}
}

func (fe *condFalseError) Error() string {
	return fmt.Sprintf("(%s) is not true\n%s", fe.cond, fe.tree)
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
		return newCondFalseError(cond, t)
	}
	return nil
}
