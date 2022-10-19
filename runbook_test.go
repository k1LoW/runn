package runn

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/tenntenn/golden"
	"gopkg.in/yaml.v2"
)

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
