package runn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/token"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tenntenn/golden"
)

func TestParseRunbook(t *testing.T) {
	es, err := os.ReadDir("testdata/book/")
	if err != nil {
		t.Fatal(err)
	}
	for _, es := range es {
		if es.IsDir() || !strings.HasSuffix(es.Name(), ".yml") {
			continue
		}
		t.Run(es.Name(), func(t *testing.T) {
			path := filepath.Join("testdata", "book", es.Name())
			f, err := os.Open(path)
			if err != nil {
				t.Error(err)
			}
			t.Cleanup(func() {
				if err := f.Close(); err != nil {
					t.Error(err)
				}
			})
			rb, err := ParseRunbook(f)
			if err != nil {
				t.Error(err)
			}
			if len(rb.Vars) == 0 && len(rb.Runners) == 0 && len(rb.Steps) == 0 {
				t.Error("want vars or runners or steps")
			}
			b, err := yaml.MarshalWithOptions(rb, encOpts...)
			if err != nil {
				t.Error(err)
			}
			rb2, err := parseRunbook(b)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(rb, rb2, cmp.AllowUnexported(runbook{})); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestAppendStep(t *testing.T) {
	tests := []struct {
		name string
		ins  [][]string
	}{
		{"curl_command", [][]string{{"curl", "https://example.com/path/to/index?foo=bar&baz=qux", "-XPOST", "-H", "Content-Type: application/json", "-d", `{"username": "alice"}`}}},
		{"grpc_command", [][]string{{"grpcurl", "-d", `{"id": 1234, "tags": ["foo","bar"]}`, "grpc.server.com:443", "my.custom.server.Service/Method"}}},
		{"exec_command", [][]string{{"echo", "hello", "world"}}},
		{"multiple_http_runner", [][]string{
			{"curl", "https://example.com/path/to/index?foo=bar&baz=qux", "-XPOST", "-H", "Content-Type: application/json", "-d", `{"username": "alice"}`},
			{"curl", "https://other.example.com/path/to/other"},
		}},
		{"multiple_exec_runner", [][]string{
			{"echo", "hello", "world"},
			{"echo", "hello", "world2"},
		}},
		{"axslog", [][]string{
			// from https://github.com/Songmu/axslogparser/blob/master/axslogparser_test.go
			{`10.0.0.11 - - [11/Jun/2017:05:56:04 +0900] "GET / HTTP/1.1" 200 741 "-" "mackerel-http-checker/0.0.1" "-"`},
			{`test.example.com 10.0.0.11 - Songmu Yaxing [11/Jun/2017:05:56:04 +0900] "GET / HTTP/1.1" 200 741`},
			{"time:08/Mar/2017:14:12:40 +0900\t" +
				"host:192.0.2.1\t" +
				"req:POST /api/v0/tsdb HTTP/1.1\t" +
				"status:200\t" +
				"size:36\t" +
				"ua:mackerel-agent/0.31.2 (Revision 775fad2)\t" +
				"reqtime:0.087\t" +
				"taken_sec:0.087\t" +
				"vhost:mackerel.io"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRunbook(tt.name)
			for _, in := range tt.ins {
				if err := rb.AppendStep(in...); err != nil {
					t.Error(err)
				}
			}

			got := new(bytes.Buffer)
			enc := yaml.NewEncoder(got, encOpts...)
			if err := enc.Encode(rb); err != nil {
				t.Error(err)
			}

			f := fmt.Sprintf("%s.append_step", tt.name)
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

func TestExpandCurlDataFiles(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name  string
		build func(path string) []string
		want  func(content string) []string
	}

	cases := []testCase{
		{
			name: "short_parameter_d",
			build: func(path string) []string {
				return []string{"curl", "-d", "@" + path, "https://example.com"}
			},
			want: func(content string) []string {
				return []string{"curl", "-d", content, "https://example.com"}
			},
		},
		{
			name: "short_parameter_d_inline"	,
			build: func(path string) []string {
				return []string{"curl", "-d@"+path, "https://example.com"}
			},
			want: func(content string) []string {
				return []string{"curl", "-d", content, "https://example.com"}
			},
		},
		{
			name: "long_parameter_data_with_space",
			build: func(path string) []string {
				return []string{"curl", "--data", "@" + path, "https://example.com"}
			},
			want: func(content string) []string {
				return []string{"curl", "--data", content, "https://example.com"}
			},
		},
		{
			name: "long_parameter_data_inline",
			build: func(path string) []string {
				return []string{"curl", "--data=@" + path, "https://example.com"}
			},
			want: func(content string) []string {
				return []string{"curl", "--data", content, "https://example.com"}
			},
		},
		{
			name: "long_parameter_data_ascii_with_space",
			build: func(path string) []string {
				return []string{"curl", "--data-ascii", "@" + path, "https://example.com"}
			},
			want: func(content string) []string {
				return []string{"curl", "--data-ascii", content, "https://example.com"}
			},
		},
		{
			name: "long_parameter_data_ascii_inline",
			build: func(path string) []string {
				return []string{"curl", "--data-ascii=@" + path, "https://example.com"}
			},
			want: func(content string) []string {
				return []string{"curl", "--data-ascii", content, "https://example.com"}
			},
		},
		{
			name: "long_parameter_data_binary_with_space",
			build: func(path string) []string {
				return []string{"curl", "--data-binary", "@" + path, "https://example.com"}
			},
			want: func(content string) []string {
				return []string{"curl", "--data-binary", content, "https://example.com"}
			},
		},
		{
			name: "long_parameter_data_binary_inline",
			build: func(path string) []string {
				return []string{"curl", "--data-binary=@" + path, "https://example.com"}
			},
			want: func(content string) []string {
				return []string{"curl", "--data-binary", content, "https://example.com"}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "payload.json")
			content := `{"message":"hello"}`
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			in := tc.build(path)
			before := append([]string(nil), in...)

			out, err := expandCurlDataFiles(in)
			if err != nil {
				t.Fatalf("expandCurlDataFiles returned error: %v", err)
			}

			want := tc.want(content)
			if diff := cmp.Diff(want, out); diff != "" {
				t.Fatalf("unexpected args (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(before, in); diff != "" {
				t.Fatalf("input slice must remain untouched (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("missing_file_returns_error", func(t *testing.T) {
		t.Parallel()

		in := []string{"curl", "-d", "@does-not-exist", "https://example.com"}
		if _, err := expandCurlDataFiles(in); err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})
}

func TestDetectRunbookAreas(t *testing.T) {
	tests := []struct {
		runbook string
		want    *areas
	}{
		{
			"testdata/book/always_failure.yml",
			&areas{
				Desc: &area{
					Start: &position{Line: 1},
					End:   &position{Line: 1},
				},
				Steps: []*area{
					{
						Start: &position{Line: 3},
						End:   &position{Line: 4},
					},
					{
						Start: &position{Line: 5},
						End:   &position{Line: 6},
					},
					{
						Start: &position{Line: 7},
						End:   &position{Line: 8},
					},
				},
			},
		},
		{
			"testdata/book/map.yml",
			&areas{
				Desc: &area{
					Start: &position{Line: 1},
					End:   &position{Line: 1},
				},
				Runners: &area{
					Start: &position{Line: 2},
					End:   &position{Line: 4},
				},
				Vars: &area{
					Start: &position{Line: 5},
					End:   &position{Line: 6},
				},
				Steps: []*area{
					{
						Start: &position{Line: 8},
						End:   &position{Line: 10},
					},
					{
						Start: &position{Line: 11},
						End:   &position{Line: 18},
					},
					{
						Start: &position{Line: 19},
						End:   &position{Line: 20},
					},
					{
						Start: &position{Line: 21},
						End:   &position{Line: 27},
					},
					{
						Start: &position{Line: 28},
						End:   &position{Line: 29},
					},
					{
						Start: &position{Line: 30},
						End:   &position{Line: 31},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.runbook, func(t *testing.T) {
			b, err := os.ReadFile(tt.runbook)
			if err != nil {
				t.Fatal(err)
			}
			got := detectRunbookAreas(string(b))
			opts := []cmp.Option{
				cmpopts.IgnoreFields(token.Position{}, "Offset"),
				cmpopts.IgnoreFields(token.Position{}, "IndentNum"),
				cmpopts.IgnoreFields(token.Position{}, "IndentLevel"),
			}
			if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestPickStepYAML(t *testing.T) {
	noColor(t)
	tests := []struct {
		runbook string
		idx     int
	}{
		{"testdata/book/http.yml", 0},
		{"testdata/book/http.yml", 8},
		{"testdata/book/github.yml", 0},
		{"testdata/book/github.yml", 3},
		{"testdata/book/github_map.yml", 0},
		{"testdata/book/github_map.yml", 3},
		{"testdata/book/single_step.yml", 0},
		{"testdata/book/single_step_map.yml", 0},
		{"testdata/book/yaml_anchor_alias.yml", 0},
		{"testdata/book/yaml_anchor_alias.yml", 7},
		{"testdata/book/yaml_anchor_alias_always_failure.yml", 0},
		{"testdata/book/yaml_anchor_alias_always_failure.yml", 1},
	}
	for _, tt := range tests {
		key := fmt.Sprintf("%s.%d", tt.runbook, tt.idx)
		t.Run(key, func(t *testing.T) {
			b, err := os.ReadFile(tt.runbook)
			if err != nil {
				t.Fatal(err)
			}
			got, err := pickStepYAML(string(b), tt.idx)
			if err != nil {
				t.Fatal(err)
			}
			f := fmt.Sprintf("pick_step.%s", filepath.Base(key))
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

func TestRunbookValidate(t *testing.T) {
	tests := []struct {
		rb      *runbook
		wantErr bool
	}{
		{&runbook{}, false},
		// labels
		{&runbook{Labels: []string{}}, false},
		{&runbook{Labels: []string{"a", "b"}}, false},
		{&runbook{Labels: []string{"key:value"}}, false},
		{&runbook{Labels: []string{"invalid+label"}}, true},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if err := tt.rb.validate(); (err != nil) != tt.wantErr {
				t.Errorf("runbook.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunbookYamlAnchorAndAlias(t *testing.T) {
	tests := []struct {
		book       string
		wantRunErr bool
	}{
		{"testdata/book/yaml_anchor_alias.yml", false},
		{"testdata/book/yaml_anchor_alias_always_failure.yml", true},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.book, func(t *testing.T) {
			t.Parallel()
			o, err := New(Book(tt.book))
			if err != nil {
				t.Errorf("got %v", err)
				return
			}

			err = o.Run(ctx)
			if err != nil {
				if !tt.wantRunErr {
					t.Errorf("got %v", err)
				}
			} else {
				if tt.wantRunErr {
					t.Errorf("want err")
				}
			}
		})
	}
}
