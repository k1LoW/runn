package runn

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/golang-sql/sqlexp/nest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/httpstub"
	"github.com/k1LoW/runn/internal/expr"
	"github.com/k1LoW/runn/internal/scope"
	"github.com/k1LoW/runn/internal/store"
	"github.com/k1LoW/runn/testutil"
	"github.com/k1LoW/stopw"
	"github.com/samber/lo"
	"github.com/tenntenn/golden"
)

var errDummy = errors.New("dummy")

var testFunc = Func("testfunc", func() string { return "this is testfunc" })

func TestExpand(t *testing.T) {
	tests := []struct {
		vars map[string]any
		in   any
		want any
	}{
		{
			map[string]any{},
			map[string]string{"key": "val"},
			map[string]any{"key": "val"},
		},
		{
			map[string]any{"one": "ichi"},
			map[string]string{"key": "{{ vars.one }}"},
			map[string]any{"key": "ichi"},
		},
		{
			map[string]any{"one": "ichi"},
			map[string]string{"{{ vars.one }}": "val"},
			map[string]any{"ichi": "val"},
		},
		{
			map[string]any{"one": 1},
			map[string]string{"key": "{{ vars.one }}"},
			map[string]any{"key": uint64(1)},
		},
		{
			map[string]any{"one": 1},
			map[string]string{"key": "{{ vars.one + 1 }}"},
			map[string]any{"key": uint64(2)},
		},
		{
			map[string]any{"one": 1},
			map[string]string{"key": "{{ string(vars.one) }}"},
			map[string]any{"key": "1"},
		},
		{
			map[string]any{"one": "01"},
			map[string]string{"path/{{ vars.one }}": "value"},
			map[string]any{"path/01": "value"},
		},
		{
			map[string]any{"year": 2022},
			map[string]string{"path?year={{ vars.year }}": "value"},
			map[string]any{"path?year=2022": "value"},
		},
		{
			map[string]any{"boolean": true},
			map[string]string{"boolean": "{{ vars.boolean }}"},
			map[string]any{"boolean": true},
		},
		{
			map[string]any{"map": map[string]any{"foo": "test", "bar": 1}},
			map[string]string{"map": "{{ vars.map }}"},
			map[string]any{"map": map[string]any{"foo": "test", "bar": uint64(1)}},
		},
		{
			map[string]any{"array": []any{map[string]any{"foo": "test1", "bar": 1}, map[string]any{"foo": "test2", "bar": 2}}},
			map[string]string{"array": "{{ vars.array }}"},
			map[string]any{"array": []any{map[string]any{"foo": "test1", "bar": uint64(1)}, map[string]any{"foo": "test2", "bar": uint64(2)}}},
		},
		{
			map[string]any{"float": float64(1)},
			map[string]string{"float": "{{ vars.float }}"},
			map[string]any{"float": uint64(1)},
		},
		{
			map[string]any{"float": float64(1.01)},
			map[string]string{"float": "{{ vars.float }}"},
			map[string]any{"float": 1.01},
		},
		{
			map[string]any{"float": float64(1.00)},
			map[string]string{"float": "{{ vars.float }}"},
			map[string]any{"float": uint64(1)},
		},
		{
			map[string]any{"float": float64(-0.9)},
			map[string]string{"float": "{{ vars.float }}"},
			map[string]any{"float": -0.9},
		},
		{
			map[string]any{"escape": "C++"},
			map[string]string{"escape": "{{ urlencode(vars.escape) }}"},
			map[string]any{"escape": "C%2B%2B"},
		},
		{
			map[string]any{"uint64": uint64(4600)},
			map[string]string{"uint64": "{{ vars.uint64 }}"},
			map[string]any{"uint64": uint64(4600)},
		},
	}
	for _, tt := range tests {
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		for k, v := range tt.vars {
			o.store.SetVar(k, v)
		}
		got, err := o.expandBeforeRecord(tt.in, &step{})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Error(diff)
		}
	}
}

func TestNewOption(t *testing.T) {
	tests := []struct {
		opts    []Option
		wantErr bool
	}{
		{
			[]Option{Book("testdata/book/book.yml"), Runner("db", "sqlite://path/to/test.db")},
			false,
		},
		{
			[]Option{Runner("db", "sqlite://path/to/test.db"), Book("testdata/book/book.yml")},
			false,
		},
		{
			[]Option{Book("testdata/book/notfound.yml")},
			true,
		},
		{
			[]Option{Runner("db", "unsupported://hostname")},
			true,
		},
		{
			[]Option{Runner("db", "sqlite://path/to/test.db"), HTTPRunner("db", "https://api.github.com", nil)},
			true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			_, err := New(tt.opts...)
			got := (err != nil)
			if got != tt.wantErr {
				t.Errorf("got %v\nwant %v", got, tt.wantErr)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/db.yml"},
		{"testdata/book/only_if_included.yml"},
		{"testdata/book/if.yml"},
		{"testdata/book/previous.yml"},
		{"testdata/book/faker.yml"},
		{"testdata/book/env.yml"},
		{"testdata/book/runner_runner.yml"},
		{"testdata/book/builtin_file.yml"},
	}
	ctx := context.Background()
	t.Setenv("DEBUG", "false")
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			_, dsn := testutil.SQLite(t)
			t.Setenv("TEST_DB_DSN", dsn)
			o, err := New(Book(tt.book), Scopes(scope.AllowRunExec), Scopes(scope.AllowReadParent))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestRunAsT(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/db.yml"},
		{"testdata/book/only_if_included.yml"},
		{"testdata/book/if.yml"},
		{"testdata/book/previous.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			db, _ := testutil.SQLite(t)
			o, err := New(Book(tt.book), DBRunner("db", db))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestRunUsingLoop(t *testing.T) {
	ts := httpstub.NewServer(t)
	counter := 0
	ts.Method(http.MethodGet).Handler(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, "%d", counter); err != nil {
			t.Fatal(err)
		}
		counter += 1
	})
	t.Cleanup(func() {
		ts.Close()
	})

	tests := []struct {
		book string
	}{
		{"testdata/book/loop.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(T(t), Book(tt.book), Runner("req", ts.Server().URL))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		paths      string
		RUNN_RUN   string
		RUNN_ID    string
		RUNN_LABEL string
		want       int
	}{
		{
			"testdata/book/**/*",
			"",
			"",
			"",
			func() int {
				e, err := os.ReadDir("testdata/book/")
				if err != nil {
					t.Fatal(err)
				}
				return len(e)
			}(),
		},
		{"testdata/book/**/*", "initdb", "", "", 1},
		{"testdata/book/**/*", "nonexistent", "", "", 0},
		{"testdata/book/**/*", "", "eb33c9aed04a7f1e03c1a1246b5d7bdaefd903d3", "", 1},
		{"testdata/book/**/*", "", "eb33c9a", "", 1},
		{"testdata/book/**/*", "", "", "http", 17},
		{"testdata/book/**/*", "", "", "openapi3", 10},
		{"testdata/book/**/*", "", "", "http,openapi3", 17},
		{"testdata/book/**/*", "", "", "http and openapi3", 10},
		{"testdata/book/**/*", "", "", "http and nothing", 0},
		{"testdata/book/**/*", "", "", "http or nothing", 17},
		{"testdata/book/**/*", "", "", "http and not openapi3", 7},
		{"testdata/book/needs_3.yml", "", "", "", 1}, // Runbooks that are only in the needs section are not counted at Load
	}

	t.Setenv("TEST_HTTP_HOST_RULE", "127.0.0.1")
	t.Setenv("TEST_GRPC_HOST_RULE", "127.0.0.1")
	t.Setenv("TEST_DB_HOST_RULE", "127.0.0.1")
	t.Setenv("TEST_SSH_HOST_RULE", "127.0.0.1")

	sshdAddr := testutil.SSHServer(t)
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Setenv("RUNN_RUN", tt.RUNN_RUN)
			t.Setenv("RUNN_ID", tt.RUNN_ID)
			t.Setenv("RUNN_LABEL", tt.RUNN_LABEL)
			opts := []Option{
				Runner("req", "https://api.github.com"),
				Runner("greq", "grpc://example.com"),
				Runner("greq2", "grpc://example.com"),
				Runner("db", "sqlite://path/to/test.db"),
				Runner("db2", "sqlite://path/to/test2.db"),
				Runner("sc", fmt.Sprintf("ssh://%s", sshdAddr)),
				Runner("sc2", fmt.Sprintf("ssh://%s", sshdAddr)),
				Runner("sc3", fmt.Sprintf("ssh://%s", sshdAddr)),
				testFunc,
			}
			ops, err := Load(tt.paths, opts...)
			if err != nil {
				t.Fatal(err)
			}
			got := len(ops.ops)
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestLoadByIDs(t *testing.T) {
	tests := []struct {
		ids []string
	}{
		{
			[]string{"8150b83", "8750a8e"},
		},
		{
			[]string{"8750a8e", "8150b83"},
		},
	}
	for _, tt := range tests {
		opts := []Option{
			RunID(tt.ids...),
		}
		ops, err := Load("testdata/book/always_*.yml", opts...)
		if err != nil {
			t.Fatal(err)
		}
		for i, id := range tt.ids {
			if !strings.HasPrefix(ops.ops[i].id, id) {
				t.Errorf("got %v\nwant %v", ops.ops[i].id, id)
			}
		}
	}
}

func TestLoadOnly(t *testing.T) {
	t.Setenv("TEST_HTTP_HOST_RULE", "127.0.0.1")
	t.Setenv("TEST_GRPC_HOST_RULE", "127.0.0.1")
	t.Setenv("TEST_DB_HOST_RULE", "127.0.0.1")
	t.Setenv("TEST_SSH_HOST_RULE", "127.0.0.1")
	t.Run("Allow to load somewhat broken runbooks", func(t *testing.T) {
		_, err := Load("testdata/book/**/*", LoadOnly())
		if err != nil {
			t.Error(err)
		}
	})
}

func TestRunN(t *testing.T) {
	tests := []struct {
		paths    string
		RUNN_RUN string
		failFast bool
		want     *runNResult
	}{
		{"testdata/book/runn_*", "", false, newRunNResult(t, 4, []*RunResult{
			{
				ID:          "84ff32ce475541124d3b28efcecb11268d79f2c6",
				Path:        "testdata/book/runn_0_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "84ff32ce475541124d3b28efcecb11268d79f2c6?step=0", Key: "0", Err: nil}},
			},
			{
				ID:          "b6d90c331b04ab198ca95b13c5f656fd2522e53b",
				Path:        "testdata/book/runn_1_fail.yml",
				Err:         errDummy,
				StepResults: []*StepResult{{ID: "b6d90c331b04ab198ca95b13c5f656fd2522e53b?step=0", Key: "0", Err: errDummy}},
			},
			{
				ID:          "faeec884c284f9c2527f840372fc01ed8351a377",
				Path:        "testdata/book/runn_2_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "faeec884c284f9c2527f840372fc01ed8351a377?step=0", Key: "0", Err: nil}},
			},
			{
				ID:          "15519f515b984b9b25dae1cfde43597cd035dc3d",
				Path:        "testdata/book/runn_3.skip.yml",
				Err:         nil,
				Skipped:     true,
				StepResults: []*StepResult{{ID: "15519f515b984b9b25dae1cfde43597cd035dc3d?step=0", Key: "0", Err: nil, Skipped: true}},
			},
		})},
		{"testdata/book/runn_*", "", true, newRunNResult(t, 4, []*RunResult{
			{
				ID:          "84ff32ce475541124d3b28efcecb11268d79f2c6",
				Path:        "testdata/book/runn_0_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "84ff32ce475541124d3b28efcecb11268d79f2c6?step=0", Key: "0", Err: nil}},
			},
			{
				ID:          "b6d90c331b04ab198ca95b13c5f656fd2522e53b",
				Path:        "testdata/book/runn_1_fail.yml",
				Err:         errDummy,
				StepResults: []*StepResult{{ID: "b6d90c331b04ab198ca95b13c5f656fd2522e53b?step=0", Key: "0", Err: errDummy}},
			},
			{
				ID:          "faeec884c284f9c2527f840372fc01ed8351a377",
				Path:        "testdata/book/runn_2_success.yml",
				Err:         nil,
				Skipped:     true,
				StepResults: []*StepResult{{ID: "faeec884c284f9c2527f840372fc01ed8351a377?step=0", Key: "0", Err: nil, Skipped: true}},
			},
			{
				ID:          "15519f515b984b9b25dae1cfde43597cd035dc3d",
				Path:        "testdata/book/runn_3.skip.yml",
				Err:         nil,
				Skipped:     true,
				StepResults: []*StepResult{{ID: "15519f515b984b9b25dae1cfde43597cd035dc3d?step=0", Key: "0", Err: nil, Skipped: true}},
			},
		})},
		{"testdata/book/runn_*", "runn_0", false, newRunNResult(t, 1, []*RunResult{
			{
				ID:          "84ff32ce475541124d3b28efcecb11268d79f2c6",
				Path:        "testdata/book/runn_0_success.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "84ff32ce475541124d3b28efcecb11268d79f2c6?step=0", Key: "0", Err: nil}},
			},
		})},
		{"testdata/book/needs_3.yml", "", false, newRunNResult(t, 3, []*RunResult{
			{
				ID:   "c4112c9cc887edf84995965b3fdd49f0b7f3424f",
				Path: "testdata/book/needs_1.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{ID: "c4112c9cc887edf84995965b3fdd49f0b7f3424f?step=0", Key: "0", Err: nil},
					{ID: "c4112c9cc887edf84995965b3fdd49f0b7f3424f?step=1", Key: "1", Err: nil},
				},
			},
			{
				ID:   "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35",
				Path: "testdata/book/needs_2.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{ID: "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35?step=0", Key: "0", Err: nil},
					{ID: "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35?step=1", Key: "1", Err: nil},
				},
			},
			{
				ID:          "727b0e891f454ff06e4a07ae441cfd7e6b2f224f",
				Path:        "testdata/book/needs_3.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "727b0e891f454ff06e4a07ae441cfd7e6b2f224f?step=0", Key: "0", Err: nil}},
			},
		})},
		{"testdata/book/needs_4.yml", "", false, newRunNResult(t, 1, []*RunResult{
			{
				ID:   "b7a2c3d17c31e57390a5bc2052be04c17d9af609",
				Path: "testdata/book/needs_4.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{
						ID:  "b7a2c3d17c31e57390a5bc2052be04c17d9af609?step=0",
						Key: "0",
						Err: nil,
						IncludedRunResults: []*RunResult{
							{
								ID:   "r-xxx",
								Path: "testdata/book/needs_1.yml",
								Err:  nil,
								StepResults: []*StepResult{
									{
										ID:  "r-xxx?step=0",
										Key: "0",
										Err: nil,
									},
									{
										ID:  "r-xxx?step=1",
										Key: "1",
										Err: nil,
									},
								},
							},
							{
								ID:   "r-xxx",
								Path: "testdata/book/needs_2.yml",
								Err:  nil,
								StepResults: []*StepResult{
									{
										ID:  "r-xxx?step=0",
										Key: "0",
										Err: nil,
									},
									{
										ID:  "r-xxx?step=1",
										Key: "1",
										Err: nil,
									},
								},
							},
							{
								ID:   "b7a2c3d17c31e57390a5bc2052be04c17d9af609?step=0",
								Path: "testdata/book/needs_3.yml",
								Err:  nil,
								StepResults: []*StepResult{
									{
										ID:  "b7a2c3d17c31e57390a5bc2052be04c17d9af609?step=0&step=0",
										Key: "0",
										Err: nil,
									},
								},
							},
						},
					},
					{
						ID:  "b7a2c3d17c31e57390a5bc2052be04c17d9af609?step=1",
						Key: "1",
						Err: nil,
					},
				},
			},
		})},
		{"testdata/book/needs_5.yml", "", false, newRunNResult(t, 2, []*RunResult{
			{
				Path: "testdata/book/needs_1.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{
						Key: "0",
						Err: nil,
					},
					{
						Key: "1",
						Err: nil,
					},
				},
			},
			{
				Path: "testdata/book/needs_5.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{
						Key: "0",
						Err: nil,
						IncludedRunResults: []*RunResult{
							{
								Path: "testdata/book/needs_2.yml",
								Err:  nil,
								StepResults: []*StepResult{
									{
										Key: "0",
										Err: nil,
									},
									{
										Key: "1",
										Err: nil,
									},
								},
							},
							{
								Path: "testdata/book/needs_3.yml",
								Err:  nil,
								StepResults: []*StepResult{
									{
										Key: "0",
										Err: nil,
									},
								},
							},
						},
					},
					{
						Key: "1",
						Err: nil,
					},
				},
			},
		})},
	}
	ctx := context.Background()
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Setenv("RUNN_RUN", tt.RUNN_RUN)
			ops, err := Load(tt.paths, FailFast(tt.failFast), testFunc)
			if err != nil {
				t.Fatal(err)
			}
			_ = ops.RunN(ctx)
			got := ops.Result().simplify()
			want := tt.want.simplify()
			opts := []cmp.Option{
				cmpopts.IgnoreFields(runResultSimplified{}, "Elapsed", "ID"),
				cmpopts.IgnoreFields(stepResultSimplified{}, "Elapsed", "ID"),
			}
			if diff := cmp.Diff(got, want, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestNeeds(t *testing.T) {
	tests := []struct {
		paths      string
		concurrent bool
		want       *runNResult
	}{
		{"testdata/book/needs_3.yml", false, newRunNResult(t, 3, []*RunResult{
			{
				ID:   "c4112c9cc887edf84995965b3fdd49f0b7f3424f",
				Path: "testdata/book/needs_1.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{ID: "c4112c9cc887edf84995965b3fdd49f0b7f3424f?step=0", Key: "0", Err: nil},
					{ID: "c4112c9cc887edf84995965b3fdd49f0b7f3424f?step=1", Key: "1", Err: nil},
				},
			},
			{
				ID:   "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35",
				Path: "testdata/book/needs_2.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{ID: "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35?step=0", Key: "0", Err: nil},
					{ID: "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35?step=1", Key: "1", Err: nil},
				},
			},
			{
				ID:          "727b0e891f454ff06e4a07ae441cfd7e6b2f224f",
				Path:        "testdata/book/needs_3.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "727b0e891f454ff06e4a07ae441cfd7e6b2f224f?step=0", Key: "0", Err: nil}},
			},
		})},
		{"testdata/book/needs_3.yml", true, newRunNResult(t, 3, []*RunResult{
			{
				ID:   "c4112c9cc887edf84995965b3fdd49f0b7f3424f",
				Path: "testdata/book/needs_1.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{ID: "c4112c9cc887edf84995965b3fdd49f0b7f3424f?step=0", Key: "0", Err: nil},
					{ID: "c4112c9cc887edf84995965b3fdd49f0b7f3424f?step=1", Key: "1", Err: nil},
				},
			},
			{
				ID:   "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35",
				Path: "testdata/book/needs_2.yml",
				Err:  nil,
				StepResults: []*StepResult{
					{ID: "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35?step=0", Key: "0", Err: nil},
					{ID: "e69056eb3a1f3f528ed41f805a35c5f7e4c1da35?step=1", Key: "1", Err: nil},
				},
			},
			{
				ID:          "727b0e891f454ff06e4a07ae441cfd7e6b2f224f",
				Path:        "testdata/book/needs_3.yml",
				Err:         nil,
				StepResults: []*StepResult{{ID: "727b0e891f454ff06e4a07ae441cfd7e6b2f224f?step=0", Key: "0", Err: nil}},
			},
		})},
	}
	ctx := context.Background()
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ops, err := Load(tt.paths, RunConcurrent(tt.concurrent, 5), testFunc)
			if err != nil {
				t.Fatal(err)
			}
			_ = ops.RunN(ctx)
			got := ops.Result().simplify()
			want := tt.want.simplify()
			opts := []cmp.Option{
				cmpopts.IgnoreFields(runResultSimplified{}, "Elapsed"),
				cmpopts.IgnoreFields(stepResultSimplified{}, "Elapsed"),
			}
			if diff := cmp.Diff(got, want, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestSkipIncluded(t *testing.T) {
	tests := []struct {
		paths        string
		skipIncluded bool
		RUNN_RUN     string
		RUNN_LABEL   string
		want         int
	}{
		{"testdata/book/include_*", false, "", "", 6},
		{"testdata/book/include_*", true, "", "", 3},
		{"testdata/book/include_*", true, "include_a.yml", "", 1},
		{"testdata/book/include_*", true, "", "label_include_a", 1},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			opts := []Option{
				SkipIncluded(tt.skipIncluded),
				RunMatch(tt.RUNN_RUN),
				RunLabel(tt.RUNN_LABEL),
				Runner("req", "https://api.github.com"),
				Runner("db", "sqlite://path/to/test.db"),
			}
			ops, err := Load(tt.paths, opts...)
			if err != nil {
				t.Fatal(err)
			}
			got := len(ops.ops)
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestSkipTest(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/skip_test.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(Book(tt.book))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestFailFast(t *testing.T) {
	tests := []struct {
		failFast    bool
		wantSuccess int
		wantFailure int
		wantSkipped int
	}{
		{false, 2, 1, 1},
		{true, 1, 1, 2},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.failFast), func(t *testing.T) {
			ops, err := Load("testdata/book/runn_*.yml", FailFast(tt.failFast))
			if err != nil {
				t.Fatal(err)
			}
			_ = ops.RunN(context.Background())

			{
				got := int(ops.Result().simplify().Success)
				if got != tt.wantSuccess {
					t.Errorf("got %v\nwant %v", got, tt.wantSuccess)
				}
			}
			{
				got := int(ops.Result().simplify().Failure)
				if got != tt.wantFailure {
					t.Errorf("got %v\nwant %v", got, tt.wantFailure)
				}
			}
			{
				got := int(ops.Result().simplify().Skipped)
				if got != tt.wantSkipped {
					t.Errorf("got %v\nwant %v", got, tt.wantSkipped)
				}
			}
		})
	}
}

func TestHookFuncTest(t *testing.T) {
	count := 0
	tests := []struct {
		book        string
		beforeFuncs []func(*RunResult) error
		afterFuncs  []func(*RunResult) error
		want        int
	}{
		{"testdata/book/skip_test.yml", nil, nil, 0},
		{
			"testdata/book/skip_test.yml",
			[]func(*RunResult) error{
				func(*RunResult) error {
					count += 3
					return nil
				},
				func(*RunResult) error {
					count = count * 2
					return nil
				},
			},
			[]func(*RunResult) error{
				func(*RunResult) error {
					count += 7
					return nil
				},
			},
			13,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		count = 0
		opts := []Option{
			Book(tt.book),
			Scopes(scope.AllowRunExec),
		}
		for _, fn := range tt.beforeFuncs {
			opts = append(opts, BeforeFunc(fn))
		}
		for _, fn := range tt.afterFuncs {
			opts = append(opts, AfterFunc(fn))
		}
		o, err := New(opts...)
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
		if count != tt.want {
			t.Errorf("got %v\nwant %v", count, tt.want)
		}
	}
}

func TestInclude(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/include_main.yml"},
		{"testdata/book/include_vars_main.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := Load(tt.book, T(t), Scopes(scope.AllowRunExec))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.RunN(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestDump(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/dump.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			dir := t.TempDir()
			out := filepath.Join(dir, "dump.out")
			opts := []Option{
				Book(tt.book),
				Stdout(io.Discard),
				Stderr(io.Discard),
				Scopes(scope.AllowRunExec),
				Var("out", out),
			}
			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
			if _, err := os.Stat(out); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestFunc(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/func.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(Book(tt.book), Func("upcase", strings.ToUpper), Stdout(io.Discard), Stderr(io.Discard))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestShard(t *testing.T) {
	tests := []struct {
		n int
	}{
		{2}, {3}, {5}, {7}, {11}, {13},
	}
	t.Setenv("TEST_HTTP_HOST_RULE", "127.0.0.1")
	t.Setenv("TEST_GRPC_HOST_RULE", "127.0.0.1")
	t.Setenv("TEST_DB_HOST_RULE", "127.0.0.1")
	sshdAddr := testutil.SSHServer(t)
	t.Setenv("TEST_SSH_HOST_RULE", sshdAddr)
	for _, tt := range tests {
		t.Run(fmt.Sprintf("n=%d", tt.n), func(t *testing.T) {
			var got []*operator
			opts := []Option{
				Runner("req", "https://api.github.com"),
				Runner("greq", "grpc://example.com"),
				Runner("greq2", "grpc://example.com"),
				Runner("db", "sqlite://path/to/test.db"),
				Runner("db2", "sqlite://path/to/test2.db"),
				Runner("sc", fmt.Sprintf("ssh://%s", sshdAddr)),
				Runner("sc2", fmt.Sprintf("ssh://%s", sshdAddr)),
				Runner("sc3", fmt.Sprintf("ssh://%s", sshdAddr)),
				Var("out", filepath.Join(t.TempDir(), "dump.out")),
				RunLabel("not needs"),
			}
			all, err := Load("testdata/book/**/*", opts...)
			if err != nil {
				t.Fatal(err)
			}
			sortOperators(all.ops)
			want := all.ops
			for i := range tt.n {
				ops, err := Load("testdata/book/**/*", append(opts, RunShard(tt.n, i))...)
				if err != nil {
					t.Fatal(err)
				}
				selected, err := ops.SelectedOperators()
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, selected...)
			}
			if len(got) != len(want) {
				t.Errorf("got %v\nwant %v", len(got), len(want))
			}
			sortOperators(got)
			allow := []any{
				operator{}, httpRunner{}, dbRunner{}, grpcRunner{}, cdpRunner{}, sshRunner{}, includeRunner{},
			}
			ignore := []any{
				step{}, store.Store{}, sql.DB{}, os.File{}, stopw.Span{}, debugger{}, nest.DB{}, Loop{}, hostRule{},
			}
			dopts := []cmp.Option{
				cmp.AllowUnexported(allow...),
				cmpopts.IgnoreUnexported(ignore...),
				cmpopts.IgnoreFields(stopw.Span{}, "ID"),
				cmpopts.IgnoreFields(operator{}, "id", "concurrency", "mu", "dbg", "needs", "nm", "maskRule", "stdout", "stderr", "deferred"),
				cmpopts.IgnoreFields(cdpRunner{}, "ctx", "cancel", "opts", "mu", "operatorID"),
				cmpopts.IgnoreFields(sshRunner{}, "client", "sess", "stdin", "stdout", "stderr", "operatorID"),
				cmpopts.IgnoreFields(grpcRunner{}, "mu", "operatorID"),
				cmpopts.IgnoreFields(dbRunner{}, "operatorID"),
				cmpopts.IgnoreFields(RunResult{}, "included", "store"),
				cmpopts.IgnoreFields(http.Client{}, "Transport"),
			}
			if diff := cmp.Diff(got, want, dopts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestVars(t *testing.T) {
	tests := []struct {
		opts    []Option
		wantErr bool
	}{
		{
			[]Option{Book("testdata/book/vars.yml"), Var("token", "world")},
			false,
		},
		{
			[]Option{Book("testdata/book/vars.yml")},
			true,
		},
		{
			[]Option{Book("testdata/book/vars_external.yml"), Var("override", "json://../vars.json")},
			false,
		},
		{
			[]Option{Book("testdata/book/vars_external.yml")},
			true,
		},
	}
	ctx := context.Background()
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			o, err := New(tt.opts...)
			if err != nil {
				t.Error(err)
			}
			if err := o.Run(ctx); err != nil {
				if !tt.wantErr {
					t.Errorf("got %v\n", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
		})
	}
}

func TestHttp(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/http.yml"},
		{"testdata/book/http_not_follow_redirect.yml"},
		{"testdata/book/http_with_json.yml"},
		{"testdata/book/http_circular_refs.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			ts := testutil.HTTPServer(t)
			t.Setenv("TEST_HTTP_ENDPOINT", ts.URL)
			o, err := New(Book(tt.book), Scopes(scope.AllowReadParent))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGrpc(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/grpc.yml"},
		{"testdata/book/grpc_with_json.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			t.Parallel()
			ts := testutil.GRPCServer(t, false, false)
			o, err := New(Book(tt.book), GrpcRunner("greq", ts.Conn()))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGrpcWithoutReflection(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/grpc.yml"},
		{"testdata/book/grpc_with_json.yml"},
	}
	ts := testutil.GRPCServer(t, true, true)
	t.Setenv("TEST_GRPC_ADDR", ts.Addr())
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			t.Parallel()
			o, err := New(Book(tt.book), Scopes(scope.AllowReadParent))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDB(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/db.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			_, dsn := testutil.SQLite(t)
			t.Setenv("TEST_DB_DSN", dsn)
			o, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAfterFuncAlwaysCall(t *testing.T) {
	tests := []struct {
		book    string
		wantErr bool
	}{
		{"testdata/book/always_success.yml", false},
		{"testdata/book/always_failure.yml", true},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			var rerr error
			called := false
			o, err := New(Book(tt.book), AfterFunc(func(rr *RunResult) error {
				if rr != nil {
					rerr = rr.Err
				}
				called = true
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			rrerr := o.Run(ctx)
			if (rrerr == nil) != (rerr == nil) {
				t.Errorf("o.Run(ctx) error: %v\nafterFunc got error:%v", rrerr, rerr)
			}
			if !called {
				t.Error("called should be true")
			}
			if rerr != nil && !tt.wantErr {
				t.Errorf("got err: %s", err)
			}
			if rerr == nil && tt.wantErr {
				t.Error("want err")
			}
		})
	}
}

func TestBeforeFuncErr(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/always_success.yml"},
		{"testdata/book/always_failure.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			o, err := New(Book(tt.book), BeforeFunc(func(*RunResult) error {
				return errors.New("before func error")
			}))
			if err != nil {
				t.Fatal(err)
			}
			if got := o.Run(ctx); got != nil {
				var be *BeforeFuncError
				if errors.Is(got, be) {
					t.Errorf("got %v\nwant %T", got, be)
				}
				return
			}
			t.Error("want err")
		})
	}
}

func TestAfterFuncErr(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/always_success.yml"},
		{"testdata/book/always_failure.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			o, err := New(Book(tt.book), AfterFunc(func(*RunResult) error {
				return errors.New("after func error")
			}))
			if err != nil {
				t.Fatal(err)
			}
			if got := o.Run(ctx); got != nil {
				var ae *AfterFuncError
				if errors.Is(got, ae) {
					t.Errorf("got %v\nwant %T", got, ae)
				}
				return
			}
			t.Error("want err")
		})
	}
}

func TestAfterFuncIf(t *testing.T) {
	tests := []struct {
		book    string
		ifCond  string
		wantErr bool
	}{
		{"testdata/book/always_success.yml", "true", true},
		{"testdata/book/always_failure.yml", "true", true},
		{"testdata/book/always_success.yml", "false", false},
		{"testdata/book/always_failure.yml", "false", true},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			o, err := New(Book(tt.book), AfterFuncIf(func(*RunResult) error {
				return errors.New("after func error")
			}, tt.ifCond))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				if !tt.wantErr {
					t.Errorf("got %v\nwant nil", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want err")
			}
		})
	}
}

func TestStoreKeys(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/store_keys.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			ts := testutil.HTTPServer(t)
			t.Setenv("TEST_HTTP_ENDPOINT", ts.URL)
			o, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestTrace(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/http.yml"},
		{"testdata/book/grpc.yml"},
		{"testdata/book/db.yml"},
	}
	ctx := context.Background()
	t.Setenv("DEBUG", "false")
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			buf := new(bytes.Buffer)
			ts := testutil.HTTPServer(t)
			t.Setenv("TEST_HTTP_ENDPOINT", ts.URL)
			_, dsn := testutil.SQLite(t)
			t.Setenv("TEST_DB_DSN", dsn)
			tg := testutil.GRPCServer(t, false, false)
			id := "1234567890"
			opts := []Option{
				Book(tt.book),
				Scopes(scope.AllowReadParent),
				GrpcRunner("greq", tg.Conn()),
				Capture(NewDebugger(buf)),
				Trace(true),
			}
			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			o.id = id
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
			if !strings.Contains(buf.String(), id) {
				t.Error("no trace")
			}
		})
	}
}

func TestLoop(t *testing.T) {
	tests := []struct {
		book    string
		count   int
		wantErr bool
	}{
		{"testdata/book/rootloop.yml", 10, false},
		{"testdata/book/rootloop.yml", 5, true},
		{"testdata/book/rootlooponly.yml", 5, false},
		{"testdata/book/rootlooponly.yml", 6, true},
	}
	ctx := context.Background()
	for i, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			key := fmt.Sprintf("testloop_count%d", i)
			got := new(bytes.Buffer)
			o, err := New(Book(tt.book), Var("lcount", tt.count), Stdout(got))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				if !tt.wantErr {
					t.Errorf("got err: %v", err)
				}
			} else {
				if tt.wantErr {
					t.Error("want err")
				}
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

func TestFailWithStepDesc(t *testing.T) {
	tests := []struct {
		book              string
		expectedSubString string
	}{
		{
			book:              "testdata/book/failure_with_step_desc.yml",
			expectedSubString: "this is description",
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			out := new(bytes.Buffer)
			opts := []Option{
				Book(tt.book),
				Stderr(out),
			}
			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			err = o.Run(ctx)
			if err == nil {
				t.Error("expected error but got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.expectedSubString) {
				t.Errorf("expected: %q is contained in result but not.\ngot string: %s", tt.expectedSubString, err.Error())
			}
		})
	}
}

func TestStepResult(t *testing.T) {
	tests := []struct {
		book  string
		force bool
		want  []*StepResult
	}{
		{"testdata/book/always_success.yml", false, []*StepResult{{Skipped: false, Err: nil}, {Skipped: false, Err: nil}, {Skipped: false, Err: nil}}},
		{"testdata/book/always_failure.yml", false, []*StepResult{{Skipped: false, Err: nil}, {Skipped: false, Err: errors.New("some error")}, {Skipped: true, Err: nil}}},
		{"testdata/book/skip_test.yml", false, []*StepResult{{Skipped: true, Err: nil}, {Skipped: false, Err: nil}}},
		{"testdata/book/only_if_included.yml", false, []*StepResult{{Skipped: true, Err: nil}, {Skipped: true, Err: nil}}},
		{"testdata/book/force.yml", false, []*StepResult{{Skipped: false, Err: nil}, {Skipped: false, Err: errors.New("some error")}, {Skipped: false, Err: nil}}},
		{"testdata/book/always_failure.yml", true, []*StepResult{{Skipped: false, Err: nil}, {Skipped: false, Err: errors.New("some error")}, {Skipped: false, Err: nil}}},
		{"testdata/book/only_if_included.yml", true, []*StepResult{{Skipped: true, Err: nil}, {Skipped: true, Err: nil}}},
		{"testdata/book/loop_and_fail.yml", true, []*StepResult{{Skipped: false, Err: nil}, {Skipped: false, Err: errors.New("some error")}, {Skipped: false, Err: nil}}},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			o, err := New(Book(tt.book), Force(tt.force), Scopes(scope.AllowRunExec))
			if err != nil {
				t.Fatal(err)
			}
			_ = o.Run(ctx)
			if len(o.runResult.StepResults) == 0 {
				t.Fatal("no step result")
			}
			for i, s := range o.steps {
				got := s.result
				if got == nil {
					t.Errorf("want step[%d] result", i)
					continue
				}
				want := tt.want[i]
				if got.Skipped != want.Skipped {
					t.Errorf("step[%d] got %v\nwant %v", i, got.Skipped, want.Skipped)
					continue
				}
				if (got.Err == nil) != (want.Err == nil) {
					t.Errorf("step[%d] got %v\nwant %v", i, got.Err, want.Err)
					continue
				}
			}
		})
	}
}

func TestStepOutcome(t *testing.T) {
	tests := []struct {
		book  string
		force bool
		want  []result
	}{
		{"testdata/book/always_success.yml", false, []result{resultSuccess, resultSuccess, resultSuccess}},
		{"testdata/book/always_failure.yml", false, []result{resultSuccess, resultFailure, resultSkipped}},
		{"testdata/book/skip_test.yml", false, []result{resultSkipped, resultSuccess}},
		{"testdata/book/only_if_included.yml", false, []result{resultSkipped, resultSkipped}},
		{"testdata/book/always_failure.yml", true, []result{resultSuccess, resultFailure, resultSuccess}},
		{"testdata/book/only_if_included.yml", true, []result{resultSkipped, resultSkipped}},
		{"testdata/book/always_success_loop.yml", false, []result{resultSuccess, resultSuccess, resultSuccess}},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			o, err := New(Book(tt.book), Force(tt.force), Scopes(scope.AllowRunExec))
			if err != nil {
				t.Fatal(err)
			}
			_ = o.Run(ctx)
			sm := o.store.ToMap()
			if o.useMap {
				if o.store.StepLen() != len(tt.want) {
					t.Errorf("got %v\nwant %v", o.store.StepLen(), len(tt.want))
				}
				i := 0
				smm, ok := sm["steps"].(map[string]map[string]any)
				if !ok {
					t.Fatal("failed to cast steps")
				}
				for _, k := range o.store.StepKeys() {
					got, ok := smm[k][store.StepKeyOutcome]
					if !ok {
						t.Error("want outcome")
						continue
					}
					want := tt.want[i]
					if got != want {
						t.Errorf("step[%d] got %v\nwant %v", i, got, want)
					}
					i++
				}
			} else {
				if o.store.StepLen() != len(tt.want) {
					t.Errorf("got %v\nwant %v", o.store.StepLen(), len(tt.want))
				}
				sl, ok := sm["steps"].([]map[string]any)
				if !ok {
					t.Fatal("failed to cast steps")
				}
				for i, s := range sl {
					got, ok := s[store.StepKeyOutcome]
					if !ok {
						t.Error("want outcome")
						continue
					}
					want := tt.want[i]
					if got != want {
						t.Errorf("step[%d] got %v\nwant %v", i, got, want)
					}
				}
			}
		})
	}
}

func TestRunnerRenew(t *testing.T) {
	book := "testdata/book/cdploop.yml"
	ctx := context.Background()
	ts := testutil.HTTPServer(t)
	var (
		o    *operator
		err  error
		once sync.Once
	)
	opts := []Option{
		Book(book),
		Var("url", ts.URL),
		BeforeFunc(func(*RunResult) error {
			once.Do(func() {
				// Close the runner connections for the first time only to get an error
				for _, r := range o.cdpRunners {
					if err := r.Close(); err != nil {
						t.Fatal(err)
					}
				}
			})
			return nil
		}),
	}
	o, err = New(opts...)
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Run(ctx); err != nil {
		t.Error(err)
	}
}

func TestTrails(t *testing.T) {
	s2 := 2
	s3 := 3

	tests := []struct {
		o    *operator
		want Trails
	}{
		{
			&operator{id: "o-a"},
			Trails{
				{Type: TrailTypeRunbook, RunbookID: "o-a"},
			},
		},
		{
			&operator{id: "o-a", parent: &step{idx: s2, key: "s-b", parent: &operator{id: "o-c"}}},
			Trails{
				{Type: TrailTypeRunbook, RunbookID: "o-c"},
				{Type: TrailTypeStep, StepIndex: &s2, StepKey: "s-b"},
				{Type: TrailTypeRunbook, RunbookID: "o-a"},
			},
		},
		{
			&operator{id: "o-a", parent: &step{idx: s2, key: "s-b", parent: &operator{id: "o-c", parent: &step{idx: s3, key: "s-d"}}}},
			Trails{
				{Type: TrailTypeStep, StepIndex: &s3, StepKey: "s-d"},
				{Type: TrailTypeRunbook, RunbookID: "o-c"},
				{Type: TrailTypeStep, StepIndex: &s2, StepKey: "s-b"},
				{Type: TrailTypeRunbook, RunbookID: "o-a"},
			},
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.o.trails()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestLabelCond(t *testing.T) {
	tests := []struct {
		runLabels []string
		labels    []string
		want      bool
	}{
		{nil, nil, true},
		{[]string{"http", "openapi3"}, []string{"http"}, true},
		{[]string{"http or openapi3"}, []string{"http"}, true},
		{[]string{"http and openapi3"}, []string{"http"}, false},
		{[]string{"user-login or user-logout"}, []string{"user-login"}, true},
		{[]string{"user:login or user:logout"}, []string{"user:login"}, true},
		{[]string{"not http"}, []string{"http"}, false},
		{[]string{"not not http"}, []string{"http"}, true},
		{[]string{"not not http"}, []string{"http"}, true},
		{[]string{"not openapi3"}, []string{"http"}, true},
		{[]string{"!http"}, []string{"http"}, false},
		{[]string{"!!http"}, []string{"http"}, true},
		{[]string{"!openapi3"}, []string{"http"}, true},
		{[]string{"http and not openapi3"}, []string{"http"}, true},
		{[]string{"http and !openapi3"}, []string{"http"}, true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.runLabels), func(t *testing.T) {
			cond := labelCond(tt.runLabels)
			got, err := expr.EvalCond(cond, labelEnv(tt.labels))
			if err != nil {
				t.Error(err)
			}
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}

func newRunNResult(t *testing.T, total int64, results []*RunResult) *runNResult {
	t.Helper()
	r := &runNResult{}
	r.Total.Store(total)
	r.RunResults = results
	return r
}

func TestRunUsingHTTPOpenAPI3(t *testing.T) {
	tests := []struct {
		skipValidateRequest string
		wantErr             bool
	}{
		{"false", true},
		{"true", false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("skipValidateRequest: %s", tt.skipValidateRequest), func(t *testing.T) {
			ctx := context.Background()
			ts := testutil.HTTPServer(t)
			t.Setenv("TEST_HTTP_ENDPOINT", ts.URL)
			t.Setenv("TEST_SKIP_VALIDATE_REQUEST", tt.skipValidateRequest)
			o, err := New(Book("testdata/book/http_skip_validate_request.yml"), Stdout(io.Discard), Stderr(io.Discard), HTTPOpenApi3s([]string{"testdata/openapi3.yml"}))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				if !tt.wantErr {
					t.Errorf("got %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("want err")
			}
		})
	}
}

func TestSortWithNeeds(t *testing.T) {
	tests := []struct {
		ops       []*operator
		sortedIds []string
		wantErr   bool
	}{
		{
			ops: func() []*operator {
				ops := tenOps(t)
				return []*operator{
					ops[0],
					ops[1],
					ops[2],
				}
			}(),
			sortedIds: []string{"id-0", "id-1", "id-2"},
			wantErr:   false,
		},
		{
			ops: func() []*operator {
				ops := tenOps(t)
				ops[1].needs = map[string]*need{
					"o2": {op: ops[2]},
				}
				return []*operator{
					ops[0],
					ops[1],
					ops[2],
				}
			}(),
			sortedIds: []string{"id-0", "id-2", "id-1"},
			wantErr:   false,
		},
		{
			ops: func() []*operator {
				ops := tenOps(t)
				ops[0].needs = map[string]*need{
					"o2": {op: ops[2]},
				}
				ops[1].needs = map[string]*need{
					"o2": {op: ops[2]},
				}
				return []*operator{
					ops[0],
					ops[1],
					ops[2],
				}
			}(),
			sortedIds: []string{"id-2", "id-0", "id-1"},
			wantErr:   false,
		},
		{
			ops: func() []*operator {
				ops := tenOps(t)
				ops[0].needs = map[string]*need{
					"o1": {op: ops[1]},
				}
				ops[1].needs = map[string]*need{
					"o2": {op: ops[2]},
				}
				return []*operator{
					ops[0],
					ops[1],
					ops[2],
				}
			}(),
			sortedIds: []string{"id-2", "id-1", "id-0"},
			wantErr:   false,
		},
		{
			ops: func() []*operator {
				ops := tenOps(t)
				ops[0].needs = map[string]*need{
					"o1": {op: ops[1]},
					"o3": {op: ops[3]},
				}
				ops[1].needs = map[string]*need{
					"o2": {op: ops[2]},
					"o3": {op: ops[3]},
				}
				ops[3].needs = map[string]*need{
					"o2": {op: ops[2]},
				}
				return []*operator{
					ops[0],
					ops[1],
					ops[2],
					ops[3],
				}
			}(),
			sortedIds: []string{"id-2", "id-3", "id-1", "id-0"},
			wantErr:   false,
		},
		{
			ops: func() []*operator {
				ops := tenOps(t)
				ops[0].needs = map[string]*need{
					"o1": {op: ops[1]},
				}
				ops[1].needs = map[string]*need{
					"o2": {op: ops[2]},
				}
				ops[2].needs = map[string]*need{
					"o0": {op: ops[0]},
				}
				return []*operator{
					ops[0],
					ops[1],
					ops[2],
					ops[3],
				}
			}(),
			sortedIds: nil,
			wantErr:   true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			sorted, err := sortWithNeeds(tt.ops)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatal(err)
			}
			if tt.wantErr {
				t.Error("want err")
				return
			}
			sortedIds := lo.Map(sorted, func(o *operator, _ int) string {
				return o.id
			})
			if diff := cmp.Diff(tt.sortedIds, sortedIds); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestStdin(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello", "hello\n"},
		{`{"hello":"world"}`, "{\n  \"hello\": \"world\"\n}\n"},
	}
	ctx := context.Background()
	book := "testdata/book/stdin.yml"
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			f, err := os.CreateTemp(t.TempDir(), "stdin")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := f.Close(); err != nil {
					t.Error(err)
				}
			})
			if _, err := f.WriteString(tt.in); err != nil {
				t.Fatal(err)
			}
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			if err := store.SetStdin(f); err != nil {
				t.Fatal(err)
			}
			buf := new(bytes.Buffer)
			opts := []Option{
				Book(book),
				Stdout(buf),
				Scopes(scope.AllowRunExec),
			}
			o, err := New(opts...)
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
			got := buf.String()
			if got != tt.want {
				t.Errorf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func tenOps(t *testing.T) []*operator {
	t.Helper()
	var ops []*operator
	for i := range 10 {
		ops = append(ops, &operator{id: fmt.Sprintf("id-%d", i)})
	}
	return ops
}
