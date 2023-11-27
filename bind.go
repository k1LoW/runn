package runn

import (
	"context"
	"fmt"

	"github.com/samber/lo"
)

const bindRunnerKey = "bind"

type bindRunner struct{}

func newBindRunner() *bindRunner {
	return &bindRunner{}
}

func (rnr *bindRunner) Run(ctx context.Context, s *step, first bool) error {
	o := s.parent
	cond := s.bindCond
	store := o.store.toMap()
	store[storeRootKeyIncluded] = o.included
	if first {
		store[storeRootPrevious] = o.store.latest()
	} else {
		store[storeRootPrevious] = o.store.previous()
		store[storeRootKeyCurrent] = o.store.latest()
	}
	for k, v := range cond {
		if lo.Contains(reservedStoreRootKeys, k) {
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
