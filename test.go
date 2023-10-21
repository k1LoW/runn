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

func newTestRunner() (*testRunner, error) {
	return &testRunner{}, nil
}

func (rnr *testRunner) Run(ctx context.Context, s *step, first bool) error {
	o := s.parent
	cond := s.testCond
	store := o.store.toMap()
	store[storeIncludedKey] = o.included
	if first {
		store[storePreviousKey] = o.store.latest()
	} else {
		store[storePreviousKey] = o.store.previous()
		store[storeCurrentKey] = o.store.latest()
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
