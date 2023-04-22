package runn

import "errors"

type step struct {
	key           string
	runnerKey     string
	desc          string
	ifCond        string
	loop          *Loop
	httpRunner    *httpRunner
	httpRequest   map[string]interface{}
	dbRunner      *dbRunner
	dbQuery       map[string]interface{}
	grpcRunner    *grpcRunner
	grpcRequest   map[string]interface{}
	cdpRunner     *cdpRunner
	cdpActions    map[string]interface{}
	sshRunner     *sshRunner
	sshCommand    map[string]interface{}
	execRunner    *execRunner
	execCommand   map[string]interface{}
	testRunner    *testRunner
	testCond      string
	dumpRunner    *dumpRunner
	dumpRequest   *dumpRequest
	bindRunner    *bindRunner
	bindCond      map[string]string
	includeRunner *includeRunner
	includeConfig *includeConfig
	// operator related to step
	parent *operator
	debug  bool
	result *StepResult
}

func newStep(key string, parent *operator) *step {
	return &step{key: key, parent: parent, debug: parent.debug}
}

func (s *step) generateID() ID {
	id := ID{
		Type:          IDTypeStep,
		Desc:          s.desc,
		StepKey:       s.key,
		StepRunnerKey: s.runnerKey,
	}
	switch {
	case s.httpRunner != nil && s.httpRequest != nil:
		id.StepRunnerType = RunnerTypeHTTP
	case s.dbRunner != nil && s.dbQuery != nil:
		id.StepRunnerType = RunnerTypeDB
	case s.grpcRunner != nil && s.grpcRequest != nil:
		id.StepRunnerType = RunnerTypeGRPC
	case s.cdpRunner != nil && s.cdpActions != nil:
		id.StepRunnerType = RunnerTypeCDP
	case s.sshRunner != nil && s.sshCommand != nil:
		id.StepRunnerType = RunnerTypeSSH
	case s.execRunner != nil && s.execCommand != nil:
		id.StepRunnerType = RunnerTypeExec
	case s.includeRunner != nil && s.includeConfig != nil:
		id.StepRunnerType = RunnerTypeInclude
	case s.dumpRunner != nil && s.dumpRequest != nil:
		id.StepRunnerType = RunnerTypeDump
	case s.bindRunner != nil && s.bindCond != nil:
		id.StepRunnerType = RunnerTypeBind
	case s.testRunner != nil && s.testCond != "":
		id.StepRunnerType = RunnerTypeTest
	}

	return id
}

func (s *step) ids() IDs {
	var ids IDs
	if s.parent != nil {
		ids = s.parent.ids()
	}
	ids = append(ids, s.generateID())
	return ids
}

func (s *step) setResult(err error) {
	if s.result != nil {
		panic("duplicate record of step results")
	}
	if errors.Is(errStepSkiped, err) {
		s.result = &StepResult{Key: s.key, Desc: s.desc, Skipped: true, Err: nil}
		return
	}
	s.result = &StepResult{Key: s.key, Desc: s.desc, Skipped: false, Err: err}
}

func (s *step) clearResult() {
	s.result = nil
}
