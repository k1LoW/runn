package builtin

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		x    any
		y    any
		want string
	}{
		{1, 1, ""},
		{1, 2, `  float64(
- 	1,
+ 	2,
  )
`},
		{1, "1", `  any(
- 	float64(1),
+ 	string("1"),
  )
`},
		{"foo", "foo", ""},
		{"foo", "bar", `  string(
- 	"foo",
+ 	"bar",
  )
`},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": true}, ""},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, `  map[string]any{
- 	"bar": bool(true),
+ 	"bar": bool(false),
  	"foo": string("1"),
  }
`},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got, err := Diff(tt.x, tt.y)
			if err != nil {
				t.Error(err)
			}
			normalizedGot := strings.ReplaceAll(got, "\u00A0", " ") // NBSP => whitespace
			if diff := cmp.Diff(tt.want, normalizedGot); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestDiffWithIgnorePathOrKeys(t *testing.T) {
	tests := []struct {
		x                any
		y                any
		ignorePathOrKeys []string
		want             string
	}{
		{1, 1, []string{"1"}, ""},
		{nil, nil, []string{"foo"}, ""},
		{nil, 1, []string{"foo"},
			`  any(
+ 	float64(1),
  )
`},
		{nil, 1, []string{"foo"},
			`  any(
+ 	float64(1),
  )
`},
		{nil, 1, []string{""},
			`  any(
+ 	float64(1),
  )
`},
		{nil, 1, []string{"."}, ""},
		{nil, 1, []string{".foo"},
			`  any(
+ 	float64(1),
  )
`},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "BAR": true}, []string{"bar"},
			`  map[string]any{
+ 	"BAR": bool(true),
  	... // 1 ignored and 1 identical entries
  }
`},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": true}, []string{}, ""},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{"bar"}, ""},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{".bar"}, ""},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{".[\"bar\"]"}, ""},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{"foo"},
			`  map[string]any{
- 	"bar": bool(true),
+ 	"bar": bool(false),
  	... // 1 ignored entry
  }
`},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{".foo"},
			`  map[string]any{
- 	"bar": bool(true),
+ 	"bar": bool(false),
  	... // 1 ignored entry
  }
`},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{"foo": "1", "bar": false}, []string{".[\"foo\"]"},
			`  map[string]any{
- 	"bar": bool(true),
+ 	"bar": bool(false),
  	... // 1 ignored entry
  }
`},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{}, []string{"foo", "bar"}, ""},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{}, []string{".foo", ".bar"}, ""},
		{map[string]any{"foo": "1", "bar": true}, map[string]any{}, []string{".[\"foo\"]", ".[\"bar\"]"}, ""},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []string{"."}, ""},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []string{".[0]"}, // NOTE: cmp.Diff() returns unintuitive output in this case
			`  []any{
  	... // 2 ignored elements
  	float64(2),
  	float64(3),
+ 	float64(4),
  }
`},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []string{".[3]"}, ""},
		{[]int{1, 2, 3}, []int{1, 2, 3, 4}, []string{".[1]", ".[3]"}, ""},
		{
			[]map[string]any{{"a": "A", "b": "B"}, {"a": "1", "b": "B", "c": "C"}},
			[]map[string]any{{"a": "A", "b": "x"}, {"a": "1", "b": "B", "c": "x"}},
			[]string{".[0].b"},
			`  []any{
  	map[string]any{"a": string("A"), ...},
  	map[string]any{
  		"a": string("1"),
  		"b": string("B"),
- 		"c": string("C"),
+ 		"c": string("x"),
  	},
  }
`,
		},
		{
			[]map[string]any{{"a": "A", "b": "B"}, {"a": "1", "b": "B", "c": "C"}},
			[]map[string]any{{"a": "A", "b": "x"}, {"a": "1", "b": "B", "c": "x"}},
			[]string{".[0].b", ".[1].c"},
			"",
		},
		{
			[]map[string]any{{"a": "A", "b": "B"}, {"a": "1", "b": "B", "c": "C"}},
			[]map[string]any{{"a": "A", "b": "x"}, {"a": "1", "b": "B", "c": "x"}},
			[]string{".[].b", ".[].c"},
			"",
		},
		{
			[]map[string]any{{"a": map[string]any{"b": map[string]any{"c": "foo", "d": true}}}},
			[]map[string]any{{"a": map[string]any{"b": map[string]any{"c": "foo", "d": false}}}},
			[]string{".. | .d?"},
			"",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			got, err := Diff(tt.x, tt.y, tt.ignorePathOrKeys...)
			if err != nil {
				t.Error(err)
			}
			normalizedGot := strings.ReplaceAll(got, "\u00A0", " ") // NBSP => whitespace
			if diff := cmp.Diff(tt.want, normalizedGot); diff != "" {
				t.Error(diff)
			}
		})
	}
}
