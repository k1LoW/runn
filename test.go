package runn

import (
	"context"
	"fmt"

	"github.com/k1LoW/runn/exprtrace"
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
	store := exprtrace.EvalEnv(o.store.toMap())
	store[storeRootKeyIncluded] = o.included
	if first {
		store[storeRootKeyPrevious] = o.store.latest()
	} else {
		store[storeRootKeyPrevious] = o.store.previous()
		store[storeRootKeyCurrent] = o.store.latest()
	}
	if err := rnr.run(ctx, cond, store, s, first); err != nil {
		return err
	}
	return nil
}

func (rnr *testRunner) run(_ context.Context, cond string, store exprtrace.EvalEnv, s *step, first bool) error {
	o := s.parent
	tf, err := EvalWithTrace(cond, store)
	if err != nil {
		return err
	}
	if !tf.OutputAsBool() {
		t, err := tf.FormatTraceTree()
		if err != nil {
			return err
		}
		return newCondFalseError(cond, t)
	}
	if first {
		o.record(nil)
	}
	return nil
}
