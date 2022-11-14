package runn

type RunResult struct {
	Desc string
	Path string
	Err  error
}

func newRunResult(desc, path string, err error) *RunResult {
	return &RunResult{
		Desc: desc,
		Path: path,
		Err:  err,
	}
}
