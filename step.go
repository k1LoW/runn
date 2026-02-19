package runn

import (
	"errors"
	"fmt"
)

type step struct {
	idx       int    // index of step in operator
	key       string // key of step in operator
	runnerKey string
	desc      string
	ifCond    string
	deferred  bool // deferred step runs after all other steps like defer in Go
	force     bool // forceed run per step
	loop      *Loop
	// loopIndex - Index of the loop is dynamically recorded at runtime
	loopIndex        *int
	httpRunner       *httpRunner
	httpRequest      map[string]any
	dbRunner         *dbRunner
	dbQuery          map[string]any
	grpcRunner       *grpcRunner
	grpcRequest      map[string]any
	cdpRunner        *cdpRunner
	cdpActions       map[string]any
	sshRunner        *sshRunner
	sshCommand       map[string]any
	execRunner       *execRunner
	execCommand      map[string]any
	testRunner       *testRunner
	testCond         string
	dumpRunner       *dumpRunner
	dumpRequest      *dumpRequest
	bindRunner       *bindRunner
	bindCond         map[string]any
	includeRunner    *includeRunner
	includeConfig    *includeConfig
	runnerRunner     *runnerRunner
	runnerDefinition map[string]any

	// runner values not yet detected.
	runnerValues map[string]any

	// operator related to step
	parent  *operator
	rawStep map[string]any
	nodes   map[string]any
	debug   bool
	result  *StepResult
}

func newStep(idx int, key string, parent *operator, rawStep map[string]any) *step {
	copied, _ := dcopy(rawStep).(map[string]any)
	return &step{idx: idx, key: key, parent: parent, rawStep: copied, debug: parent.debug}
}

// expandNodes expands the nodes of the step using store at the moment of the call.
func (s *step) expandNodes() (map[string]any, error) {
	if s.nodes != nil {
		return s.nodes, nil
	}
	o := s.parent
	nodes, err := o.expandBeforeRecord(s.rawStep, s)
	if err != nil {
		return nil, err
	}
	var ok bool
	s.nodes, ok = nodes.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("step %s: invalid nodes type: %T", s.key, nodes)
	}
	return s.nodes, nil
}

func (s *step) runnerType() RunnerType {
	switch {
	case s.httpRunner != nil && s.httpRequest != nil:
		return RunnerTypeHTTP
	case s.dbRunner != nil && s.dbQuery != nil:
		return RunnerTypeDB
	case s.grpcRunner != nil && s.grpcRequest != nil:
		return RunnerTypeGRPC
	case s.cdpRunner != nil && s.cdpActions != nil:
		return RunnerTypeCDP
	case s.sshRunner != nil && s.sshCommand != nil:
		return RunnerTypeSSH
	case s.execRunner != nil && s.execCommand != nil:
		return RunnerTypeExec
	case s.includeRunner != nil && s.includeConfig != nil:
		return RunnerTypeInclude
	case s.dumpRunner != nil && s.dumpRequest != nil:
		return RunnerTypeDump
	case s.bindRunner != nil && s.bindCond != nil:
		return RunnerTypeBind
	case s.testRunner != nil && s.testCond != "":
		return RunnerTypeTest
	case s.runnerRunner != nil && s.runnerDefinition != nil:
		return RunnerTypeRunner
	default:
		return ""
	}
}

func (s *step) generateTrail() Trail {
	return Trail{
		Type:            TrailTypeStep,
		Desc:            s.desc,
		StepIndex:       &s.idx,
		StepKey:         s.key,
		StepRunnerKey:   s.runnerKey,
		StepRunnerType:  s.runnerType(),
	}
}

// runbookID returns id of the root runbook.
func (s *step) runbookID() string { //nolint:unused
	return s.trails().runbookID()
}

func (s *step) trails() Trails {
	var trs Trails
	if s.parent != nil {
		trs = s.parent.trails()
	}
	trs = append(trs, s.generateTrail())
	if s.loopIndex != nil {
		trs = append(trs, Trail{
			Type:          TrailTypeLoop,
			LoopIndex:     s.loopIndex,
			StepIndex:     &s.idx,
			StepKey:       s.key,
			StepRunnerKey: s.runnerKey,
		})
	}
	return trs
}

func (s *step) setResult(err error) {
	if s.result != nil {
		panic("duplicate record of step results")
	}
	var runResults []*RunResult
	if s.includeRunner != nil {
		runResults = s.includeRunner.runResults
	}
	if errors.Is(errStepSkipped, err) {
		s.result = &StepResult{ID: s.runbookID(), Index: s.idx, Key: s.key, Desc: s.desc, RunnerType: s.runnerType(), RunnerKey: s.runnerKey, Skipped: true, Err: nil, IncludedRunResults: runResults}
		return
	}
	s.result = &StepResult{ID: s.runbookID(), Index: s.idx, Key: s.key, Desc: s.desc, RunnerType: s.runnerType(), RunnerKey: s.runnerKey, Skipped: false, Err: err, IncludedRunResults: runResults}
}

func (s *step) clearResult() {
	s.result = nil
	s.nodes = nil
}

func (s *step) notYetDetectedRunner() bool {
	return s.httpRunner == nil &&
		s.dbRunner == nil &&
		s.grpcRunner == nil &&
		s.cdpRunner == nil &&
		s.sshRunner == nil &&
		s.execRunner == nil &&
		len(s.runnerValues) > 0
}
