package runn

import (
	"context"
	"fmt"
)

const bindRunnerKey = "bind"

type bindRunner struct{}

func newBindRunner() (*bindRunner, error) {
	return &bindRunner{}, nil
}

func (rnr *bindRunner) Run(ctx context.Context, s *step, first bool) error {
	o := s.parent
	cond := s.bindCond
	store := o.store.toMap()
	store[storeIncludedKey] = o.included
	if first {
		store[storePreviousKey] = o.store.latest()
	} else {
		store[storePreviousKey] = o.store.previous()
		store[storeCurrentKey] = o.store.latest()
	}
	for k, v := range cond {
		if k == storeVarsKey || k == storeStepsKey || k == storeParentKey || k == storeIncludedKey || k == storeCurrentKey || k == storePreviousKey || k == loopCountVarKey {
			return fmt.Errorf("%q is reserved", k)
		}
		vv, err := EvalAny(v, store)
		if err != nil {
			return err
		}
		o.store.bindVars[k] = vv
	}
	if first {
		o.record(nil)
	}
	return nil
}
