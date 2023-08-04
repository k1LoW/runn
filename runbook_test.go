package runn

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml/token"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/tenntenn/golden"
	"gopkg.in/yaml.v2"
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
			b, err := yaml.Marshal(rb)
			if err != nil {
				t.Error(err)
			}
			rb2, err := parseRunbook(b)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(rb, rb2, cmp.AllowUnexported(runbook{})); diff != "" {
				t.Errorf("%s", diff)
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
			enc := yaml.NewEncoder(got)
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

func TestDetectRunbookAreas(t *testing.T) {
	tests := []struct {
		runbook string
		want    *areas
	}{
		{
			"testdata/book/always_failure.yml",
			&areas{
				Desc: &area{
					Start: &token.Position{Line: 1, Column: 1},
					End:   &token.Position{Line: 1, Column: 7},
				},
				Steps: []*area{
					{
						Start: &token.Position{Line: 3, Column: 3},
						End:   &token.Position{Line: 4, Column: 11},
					},
					{
						Start: &token.Position{Line: 5, Column: 3},
						End:   &token.Position{Line: 6, Column: 11},
					},
					{
						Start: &token.Position{Line: 7, Column: 3},
						End:   &token.Position{Line: 8, Column: 11},
					},
				},
			},
		},
		{
			"testdata/book/map.yml",
			&areas{
				Desc: &area{
					Start: &token.Position{Line: 1, Column: 1},
					End:   &token.Position{Line: 1, Column: 7},
				},
				Runners: &area{
					Start: &token.Position{Line: 2, Column: 1},
					End:   &token.Position{Line: 4, Column: 7},
				},
				Vars: &area{
					Start: &token.Position{Line: 5, Column: 1},
					End:   &token.Position{Line: 6, Column: 13},
				},
				Steps: []*area{
					{
						Start: &token.Position{Line: 8, Column: 3},
						End:   &token.Position{Line: 10, Column: 14},
					},
					{
						Start: &token.Position{Line: 11, Column: 3},
						End:   &token.Position{Line: 18, Column: 25},
					},
					{
						Start: &token.Position{Line: 19, Column: 3},
						End:   &token.Position{Line: 20, Column: 11},
					},
					{
						Start: &token.Position{Line: 21, Column: 3},
						End:   &token.Position{Line: 27, Column: 17},
					},
					{
						Start: &token.Position{Line: 28, Column: 3},
						End:   &token.Position{Line: 29, Column: 11},
					},
					{
						Start: &token.Position{Line: 30, Column: 3},
						End:   &token.Position{Line: 31, Column: 11},
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
			got, _ := detectRunbookAreas(string(b))
			opts := []cmp.Option{
				cmpopts.IgnoreFields(token.Position{}, "Offset"),
				cmpopts.IgnoreFields(token.Position{}, "IndentNum"),
				cmpopts.IgnoreFields(token.Position{}, "IndentLevel"),
			}
			if diff := cmp.Diff(tt.want, got, opts...); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
