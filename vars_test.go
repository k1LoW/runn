package runn

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEvaluateSchema(t *testing.T) {
	brokenJson, _ := os.CreateTemp("", "broken_json")
	defer os.Remove(brokenJson.Name())
	wd, _ := os.Getwd()

	tests := []struct {
		value interface{}
		store map[string]interface{}
		want  interface{}
		error bool
	}{
		{1, nil, 1, false},
		{[]string{"1"}, nil, []string{"1"}, false},
		{"string", nil, "string", false},
		{"json://testdata/vars.json", nil, map[string]interface{}{"foo": "test", "bar": float64(1)}, false},
		{"json://not_exists.json", nil, "json://not_exists.json", true},
		{"json://" + brokenJson.Name(), nil, "json://" + brokenJson.Name(), true},
		{
			"json://testdata/non_template.json",
			map[string]interface{}{"vars": map[string]interface{}{"foo": "test", "bar": 1}},
			map[string]interface{}{"foo": "{{.vars.foo -}}", "bar": float64(1)},
			false,
		},
		{
			"json://testdata/template.json.template",
			map[string]interface{}{"vars": map[string]interface{}{"foo": "test", "bar": 1}},
			map[string]interface{}{"foo": "test", "bar": float64(1)},
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
