package runn

import (
	"context"
	"fmt"
	"sort"

	"github.com/k1LoW/runn/internal/store"
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
	sm := o.store.ToMap()
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
	keys := lo.Keys(cond)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		v := cond[k]
		if err := o.store.RecordBindVar(k, v, sm); err != nil {
			return fmt.Errorf("failed to record bind vars: %w", err)
		}
	}
	if first {
		o.record(s.idx, nil)
	}
	return nil
}
