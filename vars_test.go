package runn

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEvaluateSchema(t *testing.T) {
	brokenJson, _ := os.CreateTemp("", "broken_json")
	defer os.Remove(brokenJson.Name())

	tests := []struct {
		value interface{}
		want  interface{}
		error bool
	}{
		{1, 1, false},
		{[]string{"1"}, []string{"1"}, false},
		{"string", "string", false},
		{"json://testdata/vars.json", map[string]interface{}{"foo": "test", "bar": float64(1)}, false},
		{"json://not_exists.json", "json://not_exists.json", true},
		{"json://" + brokenJson.Name(), "json://" + brokenJson.Name(), true},
	}
	for _, tt := range tests {
		got, err := evaluateSchema(tt.value)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
		if (tt.error && err == nil) || (!tt.error && err != nil) {
			t.Errorf("no much error")
		}
	}
}
