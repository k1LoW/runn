package runn

import (
	"fmt"
	"strings"
)

type TrailType string

const (
	TrailTypeRunbook    TrailType = "runbook"
	TrailTypeStep       TrailType = "step"
	TrailTypeBeforeFunc TrailType = "beforeFunc"
	TrailTypeAfterFunc  TrailType = "afterFunc"
	TrailTypeLoop       TrailType = "loop"
)

type RunnerType string

const (
	RunnerTypeHTTP    RunnerType = "http"
	RunnerTypeDB      RunnerType = "db"
	RunnerTypeGRPC    RunnerType = "grpc"
	RunnerTypeCDP     RunnerType = "cdp"
	RunnerTypeSSH     RunnerType = "ssh"
	RunnerTypeExec    RunnerType = "exec"
	RunnerTypeTest    RunnerType = "test"
	RunnerTypeDump    RunnerType = "dump"
	RunnerTypeInclude RunnerType = "include"
	RunnerTypeBind    RunnerType = "bind"
)

// Trail - The trail of elements in the runbook at runtime.
// Trail does not use slices to copy values.
type Trail struct {
	Type           TrailType  `json:"type"`
	Desc           string     `json:"desc,omitempty"`
	RunbookID      string     `json:"id,omitempty"`
	RunbookPath    string     `json:"path,omitempty"`
	StepIndex      *int       `json:"step_index,omitempty"`
	StepKey        string     `json:"step_key,omitempty"`
	StepRunnerType RunnerType `json:"step_runner_type,omitempty"`
	StepRunnerKey  string     `json:"step_runner_key,omitempty"`
	FuncIndex      *int       `json:"func_index,omitempty"`
	LoopIndex      *int       `json:"loop_index,omitempty"`
}

type Trails []Trail

func (tr Trail) String() string { //nostyle:recvtype
	switch tr.Type {
	case TrailTypeRunbook:
		return fmt.Sprintf("runbook[%s]", tr.RunbookPath)
	case TrailTypeStep:
		return fmt.Sprintf("steps[%s]", tr.StepKey)
	case TrailTypeBeforeFunc:
		return fmt.Sprintf("beforeFunc[%d]", *tr.FuncIndex)
	case TrailTypeAfterFunc:
		return fmt.Sprintf("afterFunc[%d]", *tr.FuncIndex)
	case TrailTypeLoop:
		return fmt.Sprintf("loop[%d]", *tr.LoopIndex)
	default:
		return "invalid"
	}
}

func (trs Trails) toProfileIDs() []any { //nostyle:recvtype
	s := make([]any, len(trs))
	for i, v := range trs {
		s[i] = v
	}
	return s
}

func (trs Trails) runbookID() string { //nostyle:recvtype
	var (
		id    string
		steps []string
	)
	for _, tr := range trs {
		switch tr.Type {
		case TrailTypeRunbook:
			if id == "" {
				id = tr.RunbookID
			}
		case TrailTypeStep:
			steps = append(steps, fmt.Sprintf("step=%d", *tr.StepIndex))
		}
	}
	if len(steps) == 0 {
		return id
	}
	return fmt.Sprintf("%s?%s", id, strings.Join(steps, "&"))
}
