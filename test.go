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

func (rnr *testRunner) Run(ctx context.Context, cond string) error {
	store := rnr.operator.store.toMap()
	store[storePreviousKey] = rnr.operator.store.previous()
	store[storeCurrentKey] = rnr.operator.store.latest()
	t, err := buildTree(cond, store)
	if err != nil {
		return err
	}
	tf, err := EvalCond(cond, store)
	if err != nil {
		return err
	}
	if !tf {
		return newCondFalseError(cond, t)
	}
	return nil
}
