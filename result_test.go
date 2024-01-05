package runn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/tenntenn/golden"
)

func TestResultOut(t *testing.T) {
	tests := []struct {
		r       *runNResult
		verbose bool
	}{
		{newRunNResult(t, 4, []*RunResult{
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_0_success.yml",
				Err:  nil,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  ErrDummy,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_2_success.yml",
				Err:  nil,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_3.skip.yml",
				Err:  nil,
			},
		}), false},
		{newRunNResult(t, 5, []*RunResult{
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_0_success.yml",
				Err:  nil,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  ErrDummy,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_2_success.yml",
				Err:  nil,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_3.skip.yml",
				Err:  nil,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/always_failure.yml",
				Err:  nil,
			},
		}), false},
		{newRunNResult(t, 2, []*RunResult{
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_0_success.yml",
				Err:  nil,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  ErrDummy,
			},
		}), false},
		{newRunNResult(t, 2, []*RunResult{
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_0_success.yml",
				Err:  nil,
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  ErrDummy,
			},
		}), true},
	}
	for i, tt := range tests {
		key := fmt.Sprintf("result_out_%d", i)
		t.Run(key, func(t *testing.T) {
			got := new(bytes.Buffer)
			if err := tt.r.Out(got, tt.verbose); err != nil {
				t.Error(err)
			}
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, "testdata", key, got)
				return
			}
			if diff := golden.Diff(t, "testdata", key, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestResultOutJSON(t *testing.T) {
	tests := []struct {
		r *runNResult
	}{
		{newRunNResult(t, 4, []*RunResult{
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_0_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: nil}},
			},
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_1_fail.yml",
				Err:         ErrDummy,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: ErrDummy}},
			},
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_2_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: nil}},
			},
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_3.skip.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: nil, Skipped: true}},
			},
		})},
		{newRunNResult(t, 5, []*RunResult{
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_0_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: nil}},
			},
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_1_fail.yml",
				Err:         ErrDummy,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: ErrDummy}},
			},
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_2_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: nil}},
			},
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_3.skip.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: nil, Skipped: true}},
			},
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/always_failure.yml",
				Err:         ErrDummy,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: ErrDummy}},
			},
		})},
		{newRunNResult(t, 2, []*RunResult{
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_0_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: nil}},
			},
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_1_fail.yml",
				Err:         ErrDummy,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: ErrDummy}},
			},
		})},
		{newRunNResult(t, 2, []*RunResult{
			{
				ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path:        "testdata/book/runn_0_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: nil}},
			},
			{
				ID:   "ab13ba1e546838ceafa17f91ab3220102f397b2e",
				Path: "testdata/book/runn_1_fail.yml",
				Err:  ErrDummy,
				StepResults: []*StepResult{{ID: "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0", Key: "0", Err: ErrDummy, IncludedRunResult: &RunResult{
					ID:          "ab13ba1e546838ceafa17f91ab3220102f397b2e?step=0",
					Path:        "testdata/book/runn_included_0_fail.yml",
					Err:         ErrDummy,
					StepResults: []*StepResult{{Key: "0", Err: ErrDummy}},
				}}},
			},
		})},
	}
	for i, tt := range tests {
		key := fmt.Sprintf("result_out_json_%d", i)
		t.Run(key, func(t *testing.T) {
			got := new(bytes.Buffer)
			if err := tt.r.OutJSON(got); err != nil {
				t.Error(err)
			}
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, "testdata", key, got)
				return
			}
			if diff := golden.Diff(t, "testdata", key, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestResultElasped(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/always_success.yml"},
		{"testdata/book/always_failure.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.book, func(t *testing.T) {
			o, err := New(Book(tt.book), Profile(true))
			if err != nil {
				t.Fatal(err)
			}
			_ = o.Run(ctx)
			result := o.Result()
			if result.Elapsed == 0 {
				t.Error("cannot measure elapsed time")
			}
		})

		t.Run(tt.book, func(t *testing.T) {
			o, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			_ = o.Run(ctx)
			result := o.Result()
			if result.Elapsed != 0 {
				t.Error("elapsed time should be zero")
			}
		})
	}
}
