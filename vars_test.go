package runn

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEvaluateSchema(t *testing.T) {
	td := t.TempDir()
	brokenJson, err := os.Create(filepath.Join(td, "broken.json"))
	if err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		value   any
		store   map[string]any
		want    any
		wantErr bool
	}{
		{1, nil, 1, false},
		{[]string{"1"}, nil, []string{"1"}, false},
		{"string", nil, "string", false},
		{"json://testdata/vars.json", nil, map[string]any{"foo": "test", "bar": float64(1)}, false},
		{"yaml://testdata/vars.yaml", nil, map[string]any{"foo": "test", "bar": uint64(1), "baz": float64(2.5)}, false},
		{"yaml://testdata/vars.yml", nil, map[string]any{"foo": "test", "bar": uint64(1), "baz": float64(2.5)}, false},
		{"json://not_exists.json", nil, "json://not_exists.json", true},
		{"json://" + brokenJson.Name(), nil, "json://" + brokenJson.Name(), true},
		{
			"json://testdata/non_template.json",
			map[string]any{"vars": map[string]any{"foo": "test", "bar": 1}},
			map[string]any{"foo": "{{.vars.foo -}}", "bar": float64(1)},
			false,
		},
		{
			"json://testdata/template.json.template",
			map[string]any{"vars": map[string]any{"foo": "test", "bar": 1}},
			map[string]any{"foo": "test", "bar": float64(1)},
			false,
		},
		{
			"yaml://testdata/template.yml.template",
			map[string]any{"vars": map[string]any{"foo": "test", "bar": 1}},
			map[string]any{"foo": "test", "bar": uint64(1)},
			false,
		},
		{
			"json://testdata/newline.json",
			map[string]any{},
			map[string]any{"foo": "abc\ndef", "bar": "abc\n\ndef"},
			false,
		},
		{
			"json://testdata/invalid_ext.js",
			map[string]any{},
			"",
			true,
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			got, err := evaluateSchema(tt.value, wd, tt.store)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %s", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error")
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}
