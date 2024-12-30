package runn

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/k1LoW/runn/internal/store"
	"github.com/samber/lo"
)

type result string

const (
	resultSuccess result = "success"
	resultFailure result = "failure"
	resultSkipped result = "skipped"
)

// RunResult is the result of a runbook run.
type RunResult struct {
	ID          string        // Runbook ID
	Desc        string        // Description of runbook
	Labels      []string      // Labels of runbook
	Path        string        // Path of runbook
	Skipped     bool          // Whether runbook run was skipped or not
	Err         error         // Error during runbook run.
	StepResults []*StepResult // Step results of runbook run
	Elapsed     time.Duration // Elapsed time of runbook run
	store       *store.Store  // Store of runbook run
	included    bool          // Whether runbook is included or not
}

// StepResult is the result of a step run.
type StepResult struct {
	ID                 string        // Runbook ID
	Key                string        // Key of step
	Desc               string        // Description of step
	Skipped            bool          // Whether step run was skipped or not
	Err                error         // Error during step run.
	IncludedRunResults []*RunResult  // Run results of runbook loaded by include runner
	Elapsed            time.Duration // Elapsed time of step run
}

type runNResult struct {
	Total      atomic.Int64
	RunResults []*RunResult
	mu         sync.Mutex
}

type runNResultSimplified struct {
	Total   int64                  `json:"total"`
	Success int64                  `json:"success"`
	Failure int64                  `json:"failure"`
	Skipped int64                  `json:"skipped"`
	Results []*runResultSimplified `json:"results"`
	Elapsed time.Duration          `json:"elapsed,omitempty"`
}

type runResultSimplified struct {
	ID      string                  `json:"id"`
	Labels  []string                `json:"labels,omitempty"`
	Path    string                  `json:"path"`
	Result  result                  `json:"result"`
	Steps   []*stepResultSimplified `json:"steps"`
	Elapsed time.Duration           `json:"elapsed,omitempty"`
}

type stepResultSimplified struct {
	ID                 string                 `json:"id"`
	Key                string                 `json:"key"`
	Result             result                 `json:"result"`
	IncludedRunResults []*runResultSimplified `json:"included_run_result,omitempty"`
	Elapsed            time.Duration          `json:"elapsed,omitempty"`
}

func newRunResult(desc string, labels []string, path string, included bool, store *store.Store) *RunResult {
	return &RunResult{
		Desc:     desc,
		Labels:   labels,
		Path:     path,
		included: included,
		store:    store,
	}
}

// HasFailure returns true if any run result has failure.
func (r *runNResult) HasFailure() bool {
	for _, rr := range r.RunResults {
		if rr.Err != nil {
			return true
		}
	}
	return false
}

func (r *runNResult) Out(out io.Writer) error {
	var ts, fs string
	_, _ = fmt.Fprintln(out, "")
	if r.HasFailure() {
		_, _ = fmt.Fprintln(out, "")
		i := 1
		var err error
		for _, rr := range r.RunResults {
			i, err = rr.outFailure(out, i)
			if err != nil {
				return err
			}
		}
	}
	_, _ = fmt.Fprintln(out, "")

	rs := r.simplify()
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
	s := r.simplify()
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

func (rr *RunResult) OutFailure(out io.Writer) error {
	_, err := rr.outFailure(out, 1)
	return err
}

func (rr *RunResult) Store() map[string]any {
	return rr.store.ToMap()
}

func (r *runNResult) simplify() runNResultSimplified {
	s := runNResultSimplified{
		Total: r.Total.Load(),
	}
	for _, rr := range r.RunResults {
		switch {
		case rr.Err != nil:
			s.Failure += 1
		case rr.Skipped:
			s.Skipped += 1
		default:
			s.Success += 1
		}
		s.Results = append(s.Results, simplifyRunResult(rr))
	}
	return s
}

func (rr *RunResult) outFailure(out io.Writer, index int) (int, error) {
	const tr = "└──"
	if rr.Err == nil {
		return index, nil
	}
	paths, indexes, errs := failedRunbookPathsAndErrors(rr)
	for ii, p := range paths {
		_, _ = fmt.Fprintf(out, "%d) %s %s\n", index, normalizePath(p[0]), cyan(rr.ID))
		for iii, pp := range p[1:] {
			_, _ = fmt.Fprintf(out, "   %s%s %s\n", strings.Repeat("    ", iii), tr, pp)
		}
		_, _ = fmt.Fprint(out, sprintMultilinef("  %s\n", "%v", red(fmt.Sprintf("Failure/Error: %s", strings.TrimRight(errs[ii].Error(), "\n")))))

		last := p[len(p)-1]
		b, err := readFile(last)
		if err != nil {
			return index, err
		}

		idx := indexes[ii]
		if idx >= 0 {
			picked, err := pickStepYAML(string(b), idx)
			if err != nil {
				return index, err
			}
			_, _ = fmt.Fprintf(out, "  Failure step (%s):\n", normalizePath(last))
			_, _ = fmt.Fprint(out, sprintMultilinef("  %s\n", "%v", picked))
			_, _ = fmt.Fprintln(out, "")
		}

		index++
	}
	return index, nil
}

func failedRunbookPathsAndErrors(rr *RunResult) ([][]string, []int, []error) {
	var (
		paths   [][]string
		indexes []int
		errs    []error
	)
	if rr.Err == nil {
		return paths, indexes, errs
	}
	for i, sr := range rr.StepResults {
		if sr.Err == nil {
			continue
		}
		if len(sr.IncludedRunResults) == 0 {
			paths = append(paths, []string{rr.Path})
			errs = append(errs, sr.Err)
			indexes = append(indexes, i)
			continue
		}
		for _, ir := range sr.IncludedRunResults {
			ps, is, es := failedRunbookPathsAndErrors(ir)
			for _, p := range ps {
				p = append([]string{rr.Path}, p...)
				paths = append(paths, p)
			}
			indexes = append(indexes, is...)
			errs = append(errs, es...)
		}
	}
	if len(paths) == 0 {
		paths = append(paths, []string{rr.Path})
		errs = append(errs, rr.Err)
		indexes = append(indexes, -1)
	}
	return paths, indexes, errs
}

func simplifyRunResult(rr *RunResult) *runResultSimplified {
	if rr == nil {
		return nil
	}
	np := normalizePath(rr.Path)
	switch {
	case rr.Err != nil:
		return &runResultSimplified{
			ID:      rr.ID,
			Path:    np,
			Result:  resultFailure,
			Steps:   simplifyStepResults(rr.StepResults),
			Elapsed: rr.Elapsed,
		}
	case rr.Skipped:
		return &runResultSimplified{
			ID:      rr.ID,
			Path:    np,
			Result:  resultSkipped,
			Steps:   simplifyStepResults(rr.StepResults),
			Elapsed: rr.Elapsed,
		}
	default:
		return &runResultSimplified{
			ID:      rr.ID,
			Path:    np,
			Result:  resultSuccess,
			Steps:   simplifyStepResults(rr.StepResults),
			Elapsed: rr.Elapsed,
		}
	}
}

func simplifyStepResults(stepResults []*StepResult) []*stepResultSimplified {
	var simplified []*stepResultSimplified
	for _, sr := range stepResults {
		switch {
		case sr.Err != nil:
			simplified = append(simplified, &stepResultSimplified{
				ID:     sr.ID,
				Key:    sr.Key,
				Result: resultFailure,
				IncludedRunResults: lo.Map(sr.IncludedRunResults, func(ir *RunResult, _ int) *runResultSimplified {
					return simplifyRunResult(ir)
				}),
				Elapsed: sr.Elapsed,
			})
		case sr.Skipped:
			simplified = append(simplified, &stepResultSimplified{
				ID:     sr.ID,
				Key:    sr.Key,
				Result: resultSkipped,
				IncludedRunResults: lo.Map(sr.IncludedRunResults, func(ir *RunResult, _ int) *runResultSimplified {
					return simplifyRunResult(ir)
				}),
				Elapsed: sr.Elapsed,
			})
		default:
			simplified = append(simplified, &stepResultSimplified{
				ID:     sr.ID,
				Key:    sr.Key,
				Result: resultSuccess,
				IncludedRunResults: lo.Map(sr.IncludedRunResults, func(ir *RunResult, _ int) *runResultSimplified {
					return simplifyRunResult(ir)
				}),
				Elapsed: sr.Elapsed,
			})
		}
	}
	return simplified
}

func sprintMultilinef(lineformat, format string, a ...any) string {
	lines := strings.Split(fmt.Sprintf(format, a...), "\n")
	var formatted string
	for _, l := range lines {
		formatted += fmt.Sprintf(lineformat, l)
	}
	return formatted
}

var (
	// root = project root path.
	root string
	once sync.Once
)

func normalizePath(p string) string {
	once.Do(func() {
		root, _ = projectRoot()
	})
	if root == "" {
		return p
	}
	abs, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return p
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return p
	}
	return rel
}

func projectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if dir == filepath.Dir(dir) {
			return "", errors.New("failed to find project root")
		}
		if _, err := os.Stat(filepath.Join(dir, ".git", "config")); err == nil {
			return dir, nil
		}
		dir = filepath.Dir(dir)
	}
}
