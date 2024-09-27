package builtin

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		x    any
		y    any
		want bool
	}{
		{1, 1, true},
		{1, 2, false},
		{1, "1", false},
		{"foo", "foo", true},
		{"foo", "bar", false},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": true}, true},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, false},
	}
	for i, tt := range tests {
		got, err := Compare(tt.x, tt.y)
		if err != nil {
			t.Error(err)
		}
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestCompareWithIgnorePathOrKeys(t *testing.T) {
	tests := []struct {
		x                any
		y                any
		ignorePathOrKeys []string
		want             bool
	}{
		{1, 1, []string{"1"}, true},
		{nil, nil, []string{"foo"}, true},
		{nil, 1, []string{"foo"}, false},
		{nil, 1, []string{"foo"}, false},
		{nil, 1, []string{""}, false},
		{nil, 1, []string{"."}, true},
		{nil, 1, []string{".foo"}, false},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "BAR": true}, []string{"bar"}, false},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": true}, []string{}, true},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{"bar"}, true},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{".bar"}, true},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{".[\"bar\"]"}, true},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{"foo"}, false},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{".foo"}, false},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{".[\"foo\"]"}, false},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{}, []string{"foo", "bar"}, true},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{}, []string{".foo", ".bar"}, true},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{}, []string{".[\"foo\"]", ".[\"bar\"]"}, true},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []string{"."}, true},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []string{".[0]"}, false},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []string{".[3]"}, true},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []string{".[1]", ".[3]"}, true},
		{
			[]map[string]any{{"a": "A", "b": "B"}, {"a": "1", "b": "B", "c": "C"}},
			[]map[string]any{{"a": "A", "b": "x"}, {"a": "1", "b": "B", "c": "x"}},
			[]string{".[0].b"},
			false,
		},
		{
			[]map[string]any{{"a": "A", "b": "B"}, {"a": "1", "b": "B", "c": "C"}},
			[]map[string]any{{"a": "A", "b": "x"}, {"a": "1", "b": "B", "c": "x"}},
			[]string{".[0].b", ".[1].c"},
			true,
		},
		{
			[]map[string]any{{"a": "A", "b": "B"}, {"a": "1", "b": "B", "c": "C"}},
			[]map[string]any{{"a": "A", "b": "x"}, {"a": "1", "b": "B", "c": "x"}},
			[]string{".[].b", ".[].c"},
			true,
		},
		{
			[]map[string]any{{"a": map[string]any{"b": map[string]any{"c": "foo", "d": true}}}},
			[]map[string]any{{"a": map[string]any{"b": map[string]any{"c": "foo", "d": false}}}},
			[]string{".. | .d?"},
			true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			got, err := Compare(tt.x, tt.y, tt.ignorePathOrKeys...)
			if err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
