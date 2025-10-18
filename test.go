package runn

import (
	"context"
	"fmt"

	"github.com/k1LoW/runn/internal/expr"
	"github.com/k1LoW/runn/internal/exprtrace"
	"github.com/k1LoW/runn/internal/store"
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
	tree := sprintMultilinef("  %s\n", "%s", fe.tree)
	return fmt.Sprintf("condition is not true\n\nCondition:\n%s", tree)
}

func newTestRunner() *testRunner {
	return &testRunner{}
}

func (rnr *testRunner) Run(ctx context.Context, s *step, first bool) error {
	o := s.parent
	cond := s.testCond
	sm := exprtrace.EvalEnv(o.store.ToMap())
	sm[store.RootKeyIncluded] = o.included
	if first {
		if !s.deferred {
			sm[store.RootKeyPrevious] = o.store.Latest()
		}
	} else {
		if !s.deferred {
			sm[store.RootKeyPrevious] = o.store.Previous()
		}
		sm[store.RootKeyCurrent] = o.store.Latest()
	}
	if err := rnr.run(ctx, cond, sm, s, first); err != nil {
		return err
	}
	return nil
}

func (rnr *testRunner) run(_ context.Context, cond string, sm exprtrace.EvalEnv, s *step, first bool) error {
	o := s.parent
	tf, err := expr.EvalWithTrace(cond, sm)
	if err != nil {
		return newErrUnrecoverable(err)
	}
	if !tf.OutputAsBool() {
		t, err := tf.FormatTraceTree()
		if err != nil {
			return newErrUnrecoverable(err)
		}
		return newCondFalseError(cond, t)
	}
	if first {
		o.record(s.idx, nil)
	}
	return nil
}
