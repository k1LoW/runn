package runn

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/tenntenn/golden"
)

func TestCmdOutCaptureResult(t *testing.T) {
	noColor(t)
	tests := []struct {
		result  *RunResult
		verbose bool
	}{
		{
			&RunResult{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  errDummy,
				StepResults: []*StepResult{{Key: "0", Err: errDummy, IncludedRunResults: []*RunResult{{
					ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
					Path:        "testdata/book/runn_included_0_fail.yml",
					Err:         errDummy,
					StepResults: []*StepResult{{Key: "0", Err: errDummy}},
					included:    true,
				}}}},
			},
			false,
		},
		{
			&RunResult{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  errDummy,
				StepResults: []*StepResult{{Key: "0", Err: errDummy, IncludedRunResults: []*RunResult{{
					ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
					Path:        "testdata/book/runn_included_0_fail.yml",
					Err:         errDummy,
					StepResults: []*StepResult{{Key: "0", Err: errDummy}},
					included:    true,
				}}}},
			},
			true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			out := new(bytes.Buffer)
			o := NewCmdOut(out, tt.verbose)
			o.CaptureResult(nil, tt.result)
			f := fmt.Sprintf("runn_cmdout_%d", i)
			got := out.String()
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, "testdata", f, got)
				return
			}
			if diff := golden.Diff(t, "testdata", f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCmdOutCaptureResultByStep(t *testing.T) {
	noColor(t)
	tests := []struct {
		result  *RunResult
		verbose bool
	}{
		{
			&RunResult{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  errDummy,
				StepResults: []*StepResult{{Key: "0", Err: errDummy, IncludedRunResults: []*RunResult{{
					ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
					Path:        "testdata/book/runn_included_0_fail.yml",
					Err:         errDummy,
					StepResults: []*StepResult{{Key: "0", Err: errDummy}},
					included:    true,
				}}}},
			},
			false,
		},
		{
			&RunResult{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  errDummy,
				StepResults: []*StepResult{{Key: "0", Err: errDummy, IncludedRunResults: []*RunResult{{
					ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
					Path:        "testdata/book/runn_included_0_fail.yml",
					Err:         errDummy,
					StepResults: []*StepResult{{Key: "0", Err: errDummy}},
					included:    true,
				}}}},
			},
			true,
		},
		{
			&RunResult{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  errDummy,
				StepResults: []*StepResult{
					{
						Key: "0",
						Err: nil,
						IncludedRunResults: []*RunResult{{
							ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
							Path:        "testdata/book/runn_included_0_success.yml",
							Err:         nil,
							StepResults: []*StepResult{{Key: "0", Err: nil}},
							included:    true,
						}},
					},
					{
						Key: "1",
						Err: errDummy,
					},
				},
			},
			true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			out := new(bytes.Buffer)
			o := NewCmdOut(out, tt.verbose)
			o.CaptureResultByStep(nil, tt.result)
			f := fmt.Sprintf("runn_cmdout_by_step%d", i)
			got := out.String()
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, "testdata", f, got)
				return
			}
			if diff := golden.Diff(t, "testdata", f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
