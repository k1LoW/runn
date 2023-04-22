package runn

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

type result string

const (
	resultSuccess result = "success"
	resultFailure result = "failure"
	resultSkipped result = "skipped"
)

type RunResult struct {
	Desc        string
	Path        string
	Skipped     bool
	Err         error
	StepResults []*StepResult
	Store       map[string]interface{}
}

type StepResult struct {
	Key     string
	Skipped bool
	Err     error
}

type runNResult struct {
	Total      atomic.Int64
	RunResults sync.Map
}

type runNResultSimplified struct {
	Total   int64             `json:"total"`
	Success int64             `json:"success"`
	Failure int64             `json:"failure"`
	Skipped int64             `json:"skipped"`
	Results map[string]result `json:"results"`
}

func newRunResult(desc, path string) *RunResult {
	return &RunResult{
		Desc: desc,
		Path: path,
	}
}

func (r *runNResult) HasFailure() bool {
	f := false
	r.RunResults.Range(func(k, v any) bool {
		rr, ok := v.(*RunResult)
		if !ok {
			return false
		}
		if rr.Err != nil {
			f = true
		}
		return true
	})
	return f
}

func (r *runNResult) Simplify() runNResultSimplified {
	s := runNResultSimplified{
		Total:   r.Total.Load(),
		Results: map[string]result{},
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
			s.Failure += 1
			s.Results[kk] = resultFailure
			return true
		}
		if rr.Skipped {
			s.Skipped += 1
			s.Results[kk] = resultSkipped
			return true
		}
		s.Success += 1
		s.Results[kk] = resultSuccess
		return true
	})
	return s
}

func (r *runNResult) Out(out io.Writer) error {
	var ts, fs string
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	rs := r.Simplify()
	if rs.Total == 1 {
		ts = fmt.Sprintf("%d scenario", rs.Total)
	} else {
		ts = fmt.Sprintf("%d scenarios", rs.Total)
	}
	ss := fmt.Sprintf("%d skipped", rs.Skipped)
	if rs.Failure == 1 {
		fs = fmt.Sprintf("%d failure", rs.Failure)
	} else {
		fs = fmt.Sprintf("%d failures", rs.Failure)
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
	s := r.Simplify()
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if _, err := out.Write(b); err != nil {
		return err
	}
	if _, err := fmt.Fprint(out, "\n"); err != nil {
		return err
	}
	return nil
}
