package builtin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		x    []any
		want map[string]any
	}{
		{
			[]any{},
			map[string]any{},
		},
		{
			[]any{map[string]any{"a": 1}},
			map[string]any{"a": 1},
		},
		{
			[]any{map[string]any{"a": 1}, map[string]any{"b": 2}},
			map[string]any{"a": 1, "b": 2},
		},
		{
			[]any{map[string]any{"a": 1, "b": 2}, map[string]any{"b": 3, "c": 4}},
			map[string]any{"a": 1, "b": 3, "c": 4},
		},
		{
			[]any{map[string]any{"a": 1, "b": 2}, map[string]any{"b": 3, "c": 4}, map[string]any{"a": 5}},
			map[string]any{"a": 5, "b": 3, "c": 4},
		},
		{
			[]any{map[string]any{"a": "foo"}, map[string]any{"b": "bar"}},
			map[string]any{"a": "foo", "b": "bar"},
		},
		{
			[]any{map[string]any{"a": true}, map[string]any{"b": false}},
			map[string]any{"a": true, "b": false},
		},
		{
			[]any{map[string]any{"a": int(1)}, map[string]any{"b": int(2)}},
			map[string]any{"a": int(1), "b": int(2)},
		},
		{
			[]any{map[string]any{"a": 1.0}, map[string]any{"b": 2.0}},
			map[string]any{"a": 1.0, "b": 2.0},
		},
		{
			[]any{map[string]any{"a": nil}, map[string]any{"b": nil}},
			map[string]any{"a": nil, "b": nil},
		},
		{
			[]any{map[string]any{"a": []string{"foo"}}, map[string]any{"b": []string{"bar"}}},
			map[string]any{"a": []string{"foo"}, "b": []string{"bar"}},
		},
		{
			[]any{map[string]any{"a": map[string]string{"foo": "bar"}}, map[string]any{"baz": map[string]string{"baz": "qux"}}},
			map[string]any{"a": map[string]string{"foo": "bar"}, "baz": map[string]string{"baz": "qux"}},
		},
	}
	for _, tt := range tests {
		got, err := Merge(tt.x...)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Error(diff)
		}
	}
}
