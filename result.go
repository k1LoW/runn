package runn

type RunResult struct {
	Err error
}

func newRunResult() *RunResult {
	return &RunResult{}
}
