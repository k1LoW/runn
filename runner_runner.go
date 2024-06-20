package runn

import (
	"context"
	"fmt"

	"github.com/samber/lo"
)

const runnerRunnerKey = "runner"

type runnerRunner struct{}

func newRunnerRunner() *runnerRunner {
	return &runnerRunner{}
}

func (rnr *runnerRunner) Run(ctx context.Context, s *step) error {
	o := s.parent
	if s.runnerDefinition == nil {
		return fmt.Errorf("runner definition is nil")
	}
	switch len(lo.Keys(s.runnerDefinition)) {
	case 0:
		return fmt.Errorf("runner definition is empty: %v", s.runnerDefinition)
	case 1:
	default:
		return fmt.Errorf("only one runner can be defined: %v", s.runnerDefinition)
	}
	e, err := o.expandBeforeRecord(s.runnerDefinition)
	if err != nil {
		return err
	}
	d, ok := e.(map[string]any)
	if !ok {
		return fmt.Errorf("runner definition: %v", e)
	}
	if err := rnr.run(ctx, d, s); err != nil {
		return err
	}
	return nil
}

func (rnr *runnerRunner) run(_ context.Context, d map[string]any, s *step) error {
	o := s.parent
	bk := newBook()
	bk.runners = d
	if err := bk.parseRunners(map[string]any{}); err != nil {
		return err
	}
	for k, r := range bk.httpRunners {
		if _, ok := o.httpRunners[k]; ok {
			return fmt.Errorf("http runner key %s is already exists", k)
		}
		o.httpRunners[k] = r
	}
	for k, r := range bk.dbRunners {
		if _, ok := o.dbRunners[k]; ok {
			return fmt.Errorf("db runner key %s is already exists", k)
		}
		o.dbRunners[k] = r
	}
	for k, r := range bk.grpcRunners {
		if _, ok := o.grpcRunners[k]; ok {
			return fmt.Errorf("grpc runner key %s is already exists", k)
		}
		o.grpcRunners[k] = r
	}
	for k, r := range bk.cdpRunners {
		if _, ok := o.cdpRunners[k]; ok {
			return fmt.Errorf("cdp runner key %s is already exists", k)
		}
		o.cdpRunners[k] = r
	}
	for k, r := range bk.sshRunners {
		if _, ok := o.sshRunners[k]; ok {
			return fmt.Errorf("ssh runner key %s is already exists", k)
		}
		o.sshRunners[k] = r
	}
	o.record(map[string]any{})
	return nil
}
