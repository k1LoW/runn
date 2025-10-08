package runn

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/runn/internal/scope"
)

func TestEvaluateSchema(t *testing.T) {
	td := t.TempDir()
	brokenJSONPath := filepath.Join(td, "broken.json")
	if err := os.WriteFile(brokenJSONPath, []byte("{]"), 0600); err != nil {
		t.Fatal(err)
	}
	validJSONPath := filepath.Join(td, "valid.json")
	if err := os.WriteFile(validJSONPath, []byte(`{"foo":"test", "bar": 1, "baz": 2.5}`), 0600); err != nil {
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
		{"json://" + brokenJSONPath, nil, "json://" + brokenJSONPath, true},
		{"json://" + validJSONPath, nil, map[string]any{"foo": "test", "bar": float64(1), "baz": float64(2.5)}, false},
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
		{"json://testdata/vars*.json", nil, []any{
			map[string]any{"foo": "test", "bar": float64(1)},
			[]any{
				map[string]any{"foo": "test1", "bar": float64(1)},
				map[string]any{"foo": "test2", "bar": float64(2)},
			},
		}, false},
		{"file://testdata/vars.json", nil, `{
    "foo": "test",
    "bar": 1
}`, false},
	}
	if err := scope.Set(scope.AllowReadParent); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := scope.Set(scope.DenyReadParent); err != nil {
			t.Fatal(err)
		}
	})
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			// t.Parallel()
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
				t.Error(diff)
			}
		})
	}
}
