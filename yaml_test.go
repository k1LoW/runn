package runn

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestYamlMarshal(t *testing.T) {
	tests := []struct {
		in   interface{}
		want interface{}
	}{
		{1, []uint8("1\n")},
		{"123", []uint8("\"123\"\n")},
		{true, []uint8("true\n")},
		{[]string{"a", "b", "c"}, []uint8("- a\n- b\n- c\n")},
		{[]int{0, 10, -99}, []uint8("- 0\n- 10\n- -99\n")},
		{map[string]string{"key": "val"}, []uint8("key: val\n")},
		{map[string]interface{}{"one": 1}, []uint8("one: 1\n")},
		{map[string]interface{}{"map": map[string]interface{}{"foo": "test", "bar": 1}}, []uint8("map:\n  bar: 1\n  foo: test\n")},
		{
			map[string]interface{}{"array": []interface{}{map[string]interface{}{"foo": "test1", "bar": 1}, map[string]interface{}{"foo": "test2", "bar": 2}}},
			[]uint8("array:\n- bar: 1\n  foo: test1\n- bar: 2\n  foo: test2\n"),
		},
	}
	for _, tt := range tests {
		got, _ := yamlMarshal(tt.in)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestYamlUnmarshal(t *testing.T) {
	tests := []struct {
		in    []byte
		want  interface{}
		error bool
	}{
		{[]uint8("1\n"), 1, false},
		{[]uint8("\"123\"\n"), "123", false},
		{[]uint8("true\n"), true, false},
		{[]uint8("- a\n- b\n- c\n"), []any{string("a"), string("b"), string("c")}, false},
		{[]uint8("- \"a\"\n- b\n- \"c\"\n"), []any{string("a"), string("b"), string("c")}, false},
		{[]uint8("- 0\n- 10\n- -99\n"), []any{int(0), int(10), int(-99)}, false},
		{[]uint8("- \"0\"\n- 10.9\n- -99\n"), []any{"0", 10.9, -99}, false},
		{[]uint8("--0\n"), "--0", false},
		{[]uint8("key: val\n"), map[any]any{string("key"): string("val")}, false},
		{[]uint8("one: 1\n"), map[any]any{"one": 1}, false},
		{[]uint8("one: \n"), map[any]any{"one": nil}, false},
		{[]uint8("one: \n-\n"), map[any]any{string("one"): []any{nil}}, false},
		{[]uint8(": :--"), nil, true},
		{
			[]uint8("map:\n  bar: 1\n  foo: test\n"),
			map[any]any{string("map"): map[any]any{string("bar"): int(1), string("foo"): string("test")}},
			false,
		},
		{
			[]uint8("array:\n- bar: 1\n  foo: test1\n- bar: 2\n  foo: test2\n"),
			map[any]any{
				string("array"): []any{
					map[any]any{string("bar"): int(1), string("foo"): string("test1")},
					map[any]any{string("bar"): int(2), string("foo"): string("test2")},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		var got interface{}
		err := yamlUnmarshal(tt.in, &got)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
		if (tt.error && err == nil) || (!tt.error && err != nil) {
			t.Errorf("no much error")
		}
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		v    interface{}
		want interface{}
	}{
		{1, 1},
		{"123", "123"},
		{"'123'", "123"},
		{"\"123\"", "123"},
		{true, true},
		{[]any{string("a"), string("b"), string("c")}, []any{string("a"), string("b"), string("c")}},
		{[]string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]any{int(0), int(10), int(-99)}, []any{int(0), int(10), int(-99)}},
		{[]int{0, 10, -99}, []int{0, 10, -99}},
		{[]any{"0", 10.9, -99}, []any{"0", 10.9, -99}},
		{[]interface{}{"1", 20.9, -100}, []any{"1", 20.9, -100}},
		{map[any]any{string("key"): string("val")}, map[string]any{"key": string("val")}},
		{map[string]interface{}{"key": "val"}, map[string]any{"key": string("val")}},
		{map[string]interface{}{"'key'": "val"}, map[string]any{"key": string("val")}},
		{map[string]interface{}{"\"key\"": "val"}, map[string]any{"key": string("val")}},
		{map[int]interface{}{0: "val"}, map[int]any{0: "val"}},
		{map[interface{}]interface{}{0: "foo", "'1'": "bar"}, map[string]any{"0": string("foo"), "1": string("bar")}},
		{map[any]any{"one": 1}, map[string]any{"one": int(1)}},
		{map[any]any{"\"one\"": nil}, map[string]any{"one": nil}},
		{map[any]any{string("one"): []any{nil}}, map[string]any{"one": []any{nil}}},
		{
			map[any]any{string("map"): map[any]any{string("bar"): int(1), string("foo"): string("test")}},
			map[string]any{"map": map[string]any{"bar": int(1), "foo": string("test")}},
		},
		{
			map[string]interface{}{"map": map[string]interface{}{"bar": 1, "foo": "test"}},
			map[string]interface{}{"map": map[string]any{"bar": int(1), "foo": string("test")}},
		},
		{
			map[any]any{
				string("array"): []any{
					map[any]any{string("bar"): int(1), string("foo"): string("test1")},
					map[any]any{string("bar"): int(2), string("foo"): string("test2")},
				},
			},
			map[string]any{
				string("array"): []any{
					map[string]any{string("bar"): int(1), string("foo"): string("test1")},
					map[string]any{string("bar"): int(2), string("foo"): string("test2")},
				},
			},
		},
		{
			map[string]interface{}{
				"array": []interface{}{
					map[string]interface{}{"bar": 1, "foo": "test1"},
					map[string]any{"bar": 2, "foo": "test2"},
				},
			},
			map[string]any{
				string("array"): []any{
					map[string]any{string("bar"): int(1), string("foo"): string("test1")},
					map[string]any{string("bar"): int(2), string("foo"): string("test2")},
				},
			},
		},
		{
			[]map[string]interface{}{
				{"bar": 1, "'foo'": "test1"},
				{"bar": 2, "\"foo\"": "test2"},
			},
			[]map[string]interface{}{
				map[string]any{string("bar"): int(1), string("foo"): string("test1")},
				map[string]any{string("bar"): int(2), string("foo"): string("test2")},
			},
		},
		{
			yaml.MapSlice{
				yaml.MapItem{Key: 1, Value: "test1"},
				yaml.MapItem{Key: "'3'", Value: true},
			},
			map[string]any{"1": string("test1"), "3": bool(true)},
		},
	}
	for _, tt := range tests {
		got := normalize(tt.v)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
