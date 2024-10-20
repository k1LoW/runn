package builtin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestOmit(t *testing.T) {
	tests := []struct {
		x    any
		keys []string
		want any
	}{
		{map[string]any{"a": int(1), "b": int(2), "c": int(3)}, []string{"b"}, map[string]any{"a": int(1), "c": int(3)}},
		{map[string]any{"a": "foo", "b": "bar", "c": "baz"}, []string{"b"}, map[string]any{"a": "foo", "c": "baz"}},
		{map[string]any{"a": true, "b": true, "c": false}, []string{"b"}, map[string]any{"a": true, "c": false}},
		{map[string]any{"a": 1.0, "b": 2.0, "c": 3.0}, []string{"b"}, map[string]any{"a": 1.0, "c": 3.0}},
		{map[string]any{"a": []string{"foo"}, "b": []string{"bar"}, "c": []string{"baz"}}, []string{"b"}, map[string]any{"a": []string{"foo"}, "c": []string{"baz"}}},
		{map[string]any{"a": map[string]any{"a": 1}, "b": map[string]any{"b": 2}, "c": map[string]any{"c": 3}}, []string{"b"}, map[string]any{"a": map[string]any{"a": 1}, "c": map[string]any{"c": 3}}},
		{map[string]any{"a": int(1), "b": int(2)}, []string{"b", "not_existing_key"}, map[string]any{"a": int(1)}},
		{map[string]any{}, []string{"not_existing_key"}, map[string]any{}},
		{map[string]any{}, []string{}, map[string]any{}},
		{
			map[string]any{"a": int(1), "b": 2.0, "c": "3", "d": true, "e": nil, "f": []int{6}, "g": map[string]any{"h": 7}, "i": nil},
			[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "not_existing_key"},
			map[string]any{},
		},
	}
	for _, tt := range tests {
		got, err := Omit(tt.x, tt.keys...)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Error(diff)
		}
	}
}
