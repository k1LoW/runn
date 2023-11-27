package runn

import (
	"context"
	"fmt"
)

const testRunnerKey = "test"

type testRunner struct{}

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
	tree := SprintMultilinef("  %s\n", "%s", fe.tree)
	return fmt.Sprintf("condition is not true\n\nCondition:\n%s", tree)
}

func newTestRunner() *testRunner {
	return &testRunner{}
}

func (rnr *testRunner) Run(ctx context.Context, s *step, first bool) error {
	o := s.parent
	cond := s.testCond
	store := o.store.toMap()
	store[storeRootKeyIncluded] = o.included
	if first {
		store[storeRootPrevious] = o.store.latest()
	} else {
		store[storeRootPrevious] = o.store.previous()
		store[storeRootKeyCurrent] = o.store.latest()
	}
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
	if first {
		o.record(nil)
	}
	return nil
}
