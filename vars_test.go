package runn

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEvaluateSchema(t *testing.T) {
	brokenJson, err := os.CreateTemp("", "broken_json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(brokenJson.Name())
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		value any
		store map[string]any
		want  any
		error bool
	}{
		{1, nil, 1, false},
		{[]string{"1"}, nil, []string{"1"}, false},
		{"string", nil, "string", false},
		{"json://testdata/vars.json", nil, map[string]any{"foo": "test", "bar": float64(1)}, false},
		{"yaml://testdata/vars.yaml", nil, map[string]any{"foo": "test", "bar": uint64(1), "baz": float64(2.5)}, false},
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
			"json://testdata/newline.json",
			map[string]any{},
			map[string]any{"foo": "abc\ndef", "bar": "abc\n\ndef"},
			false,
		},
	}
	for _, tt := range tests {
		got, err := evaluateSchema(tt.value, wd, tt.store)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
		if (tt.error && err == nil) || (!tt.error && err != nil) {
			t.Errorf("no much error")
		}
	}
}
