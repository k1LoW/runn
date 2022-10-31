package runn

import "fmt"

type IDType string

const (
	IDTypeRunbook    IDType = "runbook"
	IDTypeStep       IDType = "step"
	IDTypeBeforeFunc IDType = "beforeFunc"
	IDTypeAfterFunc  IDType = "afterFunc"
)

type RunnerType string

const (
	RunnerTypeHTTP    RunnerType = "http"
	RunnerTypeDB      RunnerType = "db"
	RunnerTypeGRPC    RunnerType = "grpc"
	RunnerTypeExec    RunnerType = "exec"
	RunnerTypeTest    RunnerType = "test"
	RunnerTypeDump    RunnerType = "dump"
	RunnerTypeInclude RunnerType = "include"
)

// ID - ID and context of each element in the runbook
type ID struct {
	Type           IDType     `json:"type"`
	Desc           string     `json:"desc,omitempty"`
	RunbookID      string     `json:"id,omitempty"`
	RunbookPath    string     `json:"path,omitempty"`
	StepKey        string     `json:"key,omitempty"`
	StepRunnerType RunnerType `json:"runner_type,omitempty"`
	StepRunnerKey  string     `json:"runner_key,omitempty"`
	FuncIndex      int        `json:"func_index,omitempty"`
}

type IDs []ID

func (id ID) String() string {
	switch id.Type {
	case IDTypeRunbook:
		return fmt.Sprintf("runbook[%s]", id.RunbookPath)
	case IDTypeStep:
		return fmt.Sprintf("steps[%s]", id.StepKey)
	case IDTypeBeforeFunc:
		return fmt.Sprintf("beforeFunc[%d]", id.FuncIndex)
	case IDTypeAfterFunc:
		return fmt.Sprintf("afterFunc[%d]", id.FuncIndex)
	default:
		return "invalid"
	}
}

func (ids IDs) toInterfaceSlice() []interface{} {
	s := make([]interface{}, len(ids))
	for i, v := range ids {
		s[i] = v
	}
	return s
}
