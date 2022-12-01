package runn

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/tenntenn/golden"
)

func TestResultOut(t *testing.T) {
	tests := []struct {
		r *runNResult
	}{
		{newRunNResult(t, 4, map[string]result{
			"testdata/book/runn_0_success.yml": resultSuccess,
			"testdata/book/runn_1_fail.yml":    resultFailure,
			"testdata/book/runn_2_success.yml": resultSuccess,
			"testdata/book/runn_3.skip.yml":    resultSuccess,
		})},
		{newRunNResult(t, 5, map[string]result{
			"testdata/book/runn_0_success.yml": resultSuccess,
			"testdata/book/runn_1_fail.yml":    resultFailure,
			"testdata/book/runn_2_success.yml": resultSuccess,
			"testdata/book/runn_3.skip.yml":    resultSuccess,
			"testdata/book/always_failure.yml": resultSuccess,
		})},
		{newRunNResult(t, 2, map[string]result{
			"testdata/book/runn_0_success.yml": resultSuccess,
			"testdata/book/runn_1_fail.yml":    resultFailure,
		})},
	}
	for i, tt := range tests {
		key := fmt.Sprintf("result_out_%d", i)
		t.Run(key, func(t *testing.T) {
			got := new(bytes.Buffer)
			if err := tt.r.Out(got); err != nil {
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
		{newRunNResult(t, 4, map[string]result{
			"testdata/book/runn_0_success.yml": resultSuccess,
			"testdata/book/runn_1_fail.yml":    resultFailure,
			"testdata/book/runn_2_success.yml": resultSuccess,
			"testdata/book/runn_3.skip.yml":    resultSuccess,
		})},
		{newRunNResult(t, 5, map[string]result{
			"testdata/book/runn_0_success.yml": resultSuccess,
			"testdata/book/runn_1_fail.yml":    resultFailure,
			"testdata/book/runn_2_success.yml": resultSuccess,
			"testdata/book/runn_3.skip.yml":    resultSuccess,
			"testdata/book/always_failure.yml": resultSuccess,
		})},
		{newRunNResult(t, 2, map[string]result{
			"testdata/book/runn_0_success.yml": resultSuccess,
			"testdata/book/runn_1_fail.yml":    resultFailure,
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
