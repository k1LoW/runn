package runn

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

type RunResult struct {
	Desc    string
	Path    string
	Skipped bool
	Err     error
}

type runNResult struct {
	Total      atomic.Int64
	Success    atomic.Int64
	Failure    atomic.Int64
	Skipped    atomic.Int64
	RunResults sync.Map
}

func newRunResult(desc, path string) *RunResult {
	return &RunResult{
		Desc: desc,
		Path: path,
	}
}

func (r *runNResult) HasFailure() bool {
	return r.Failure.Load() > 0
}

func (r *runNResult) Out(out io.Writer) error {
	var ts, fs string
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	if r.Total.Load() == 1 {
		ts = fmt.Sprintf("%d scenario", r.Total.Load())
	} else {
		ts = fmt.Sprintf("%d scenarios", r.Total.Load())
	}
	ss := fmt.Sprintf("%d skipped", r.Skipped.Load())
	if r.Failure.Load() == 1 {
		fs = fmt.Sprintf("%d failure", r.Failure.Load())
	} else {
		fs = fmt.Sprintf("%d failures", r.Failure.Load())
	}
	if r.HasFailure() {
		if _, err := fmt.Fprintf(out, red("%s, %s, %s\n"), ts, ss, fs); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(out, green("%s, %s, %s\n"), ts, ss, fs); err != nil {
			return err
		}
	}
	return nil
}

func (r *runNResult) OutJSON(out io.Writer) error {
	const (
		resultSuccess = "success"
		resultFailure = "failure"
		resultSkipped = "skipped"
	)
	s := struct {
		Total   int64             `json:"total"`
		Success int64             `json:"success"`
		Failure int64             `json:"failure"`
		Skipped int64             `json:"skipped"`
		Results map[string]string `json:"results"`
	}{
		Total:   r.Total.Load(),
		Success: r.Success.Load(),
		Failure: r.Failure.Load(),
		Skipped: r.Skipped.Load(),
		Results: map[string]string{},
	}
	r.RunResults.Range(func(k, v any) bool {
		rr, ok := v.(*RunResult)
		if !ok {
			return false
		}
		kk, ok := k.(string)
		if !ok {
			return false
		}
		if rr.Err != nil {
			s.Results[kk] = resultFailure
			return true
		}
		if rr.Skipped {
			s.Results[kk] = resultSkipped
			return true
		}
		s.Results[kk] = resultSuccess
		return true
	})

	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if _, err := out.Write(b); err != nil {
		return err
	}
	return nil
}
