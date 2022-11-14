package runn

import (
	"sync"
	"sync/atomic"
)

type RunResult struct {
	Desc string
	Path string
	Err  error
}

type runNResult struct {
	Total      atomic.Int64
	Success    atomic.Int64
	Failed     atomic.Int64
	Skipped    atomic.Int64
	RunResults sync.Map
}

func newRunResult(desc, path string) *RunResult {
	return &RunResult{
		Desc: desc,
		Path: path,
	}
}
