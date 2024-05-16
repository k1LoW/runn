package runn

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/runn/builtin"
	"github.com/k1LoW/runn/exprtrace"
)

func TestEvalWithTraceFormatTraceTree(t *testing.T) {
	tests := []struct {
		in   string
		env  exprtrace.EvalEnv
		want string
	}{
		{
			"vars.key == 'hello'",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`vars.key == 'hello'
│
├── vars.key => "hello"
└── "hello"
`,
		},
		{
			"vars['key'] == 'hello'",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`vars['key'] == 'hello'
│
├── vars.key => "hello"
└── "hello"
`,
		},
		{
			`// comment
1 == 1`,
			exprtrace.EvalEnv{},
			`// comment
1 == 1
│
├── 1
└── 1
`,
		},
		{
			`1 == 1 // comment`,
			exprtrace.EvalEnv{},
			`1 == 1 // comment
│
├── 1
└── 1
`,
		},
		{
			`/* comment */ 1 == 1`,
			exprtrace.EvalEnv{},
			`/* comment */ 1 == 1
│
├── 1
└── 1
`,
		},
		{
			`1 == /* comment */ 1`,
			exprtrace.EvalEnv{},
			`1 == /* comment */ 1
│
├── 1
└── 1
`,
		},
		{
			`map([1, 2, 3], {
# * 2
}) == [2, 4, 6]`,
			exprtrace.EvalEnv{},
			`map([1, 2, 3], {
# * 2
}) == [2, 4, 6]
│
├── map([1, 2, 3], # * 2) => [2,4,6]
│   ├── [1, 2, 3] => [1,2,3]
│   └── # * 2 => [...]
│       ├── (0) ... => 2
│       │   └── # * 2 => 2
│       │       ├── # => 1
│       │       └── 2
│       ├── (1) ... => 4
│       │   └── # * 2 => 4
│       │       ├── # => 2
│       │       └── 2
│       └── (2) ... => 6
│           └── # * 2 => 6
│               ├── # => 3
│               └── 2
└── [2, 4, 6] => [2,4,6]
`,
		},
		{
			"arr[:2] == [1,2]",
			exprtrace.EvalEnv{
				"arr": []int{1, 2, 3, 4, 5},
			},
			`arr[:2] == [1,2]
│
├── arr[:2] => [1,2]
│   ├── arr => [1,2,3,4,5]
│   ├── (from) [not specified]
│   └── (to) 2
└── [1, 2] => [1,2]
`,
		},
		{
			"printf('%s world', vars.key) == 'hello world'",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"key": "hello",
				},
				"printf": func(format string, a ...any) string {
					return fmt.Sprintf(format, a...)
				},
			},
			`printf('%s world', vars.key) == 'hello world'
│
├── printf("%s world", vars.key) => "hello world"
│   ├── "%s world"
│   └── vars.key => "hello"
└── "hello world"
`,
		},
		{
			"vars.tests[i] == vars.wants[i]",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"tests": []int{1, 2, 3},
					"wants": []int{0, 1, 2},
				},
				"i": 1,
			},
			`vars.tests[i] == vars.wants[i]
│
├── vars.tests[i] => 2
└── vars.wants[i] => 1
`,
		},
		{
			"vars.expected in map(vars.res, { #.value })",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"expected": 10,
					"res":      []any{map[string]any{"value": 10}, map[string]any{"value": 20}, map[string]any{"value": 30}},
				},
			},
			`vars.expected in map(vars.res, { #.value })
│
├── vars.expected => 10
└── map(vars.res, .value) => [10,20,30]
    ├── vars.res => [{"value":10},{"value":20},{"value":30}]
    └── .value => [...]
        ├── (0) ... => 10
        │   └── .value => 10
        ├── (1) ... => 20
        │   └── .value => 20
        └── (2) ... => 30
            └── .value => 30
`,
		},
		{
			"diff(intersect(vars.v1, vars.v2), vars.v3) == ''",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"v1": []any{1, 2, 3, 4},
					"v2": []any{1, 3, 5, 7},
					"v3": []any{1, 3},
				},
				"diff":      builtin.Diff,
				"intersect": builtin.Intersect,
			},
			`diff(intersect(vars.v1, vars.v2), vars.v3) == ''
│
├── diff(intersect(vars.v1, vars.v2), vars.v3) => ""
│   ├── intersect(vars.v1, vars.v2) => [1,3]
│   │   ├── vars.v1 => [1,2,3,4]
│   │   └── vars.v2 => [1,3,5,7]
│   └── vars.v3 => [1,3]
└── ""
`,
		},
		{
			`vars.key == "hello\nworld"`,
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`vars.key == "hello\nworld"
│
├── vars.key => "hello"
└── "hello\nworld"
`,
		},
		{
			"vars.key == -100",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"key": 100,
				},
			},
			`vars.key == -100
│
├── vars.key => 100
└── -100
`,
		},
		{
			"true && false && (true || true)",
			exprtrace.EvalEnv{},
			`true && false && (true || true)
│
├── true && false => false
│   ├── true
│   └── false
└── true || true => [not evaluated]
`,
		},
		{
			"user.Name | lower() | split(\" \") == [\"foo\", \"bar\"]",
			exprtrace.EvalEnv{
				"user": map[string]any{
					"Name": "Foo Bar",
				},
			},
			`user.Name | lower() | split(" ") == ["foo", "bar"]
│
├── split(lower(user.Name), " ") => ["foo","bar"]
│   ├── lower(user.Name) => "foo bar"
│   │   └── user.Name => "Foo Bar"
│   └── " "
└── ["foo", "bar"] => ["foo","bar"]
`,
		},
		{
			"1..3 == [1, 2, 3]",
			exprtrace.EvalEnv{},
			`1..3 == [1, 2, 3]
│
├── 1..3 => [1,2,3]
│   ├── 1
│   └── 3
└── [1, 2, 3] => [1,2,3]
`,
		},
		{
			"let x = 42; x * 2",
			exprtrace.EvalEnv{},
			`let x = 42; x * 2
│
├── let x = 42
└── x * 2 => 84
    ├── x => 42
    └── 2
`,
		},
		{
			"trim(\"__Hello__\", \"_\") == \"Hello\"",
			exprtrace.EvalEnv{},
			`trim("__Hello__", "_") == "Hello"
│
├── trim("__Hello__", "_") => "Hello"
│   ├── "__Hello__"
│   └── "_"
└── "Hello"
`,
		},
		{
			"findIndex([1, 2, 3, 4], # > 2) == 2",
			exprtrace.EvalEnv{},
			`findIndex([1, 2, 3, 4], # > 2) == 2
│
├── findIndex([1, 2, 3, 4], # > 2) => 2
│   ├── [1, 2, 3, 4] => [1,2,3,4]
│   └── # > 2 => [...]
│       ├── (0) ... => false
│       │   └── # > 2 => false
│       │       ├── # => 1
│       │       └── 2
│       ├── (1) ... => false
│       │   └── # > 2 => false
│       │       ├── # => 2
│       │       └── 2
│       └── (2) ... => true
│           └── # > 2 => true
│               ├── # => 3
│               └── 2
└── 2
`,
		},

		{
			"map([1, 2, 3], { map(1..#, {#+#}) })",
			exprtrace.EvalEnv{},
			`map([1, 2, 3], { map(1..#, {#+#}) })
│
├── [1, 2, 3] => [1,2,3]
└── map(1..#, # + #) => [...]
    ├── (0) ... => [2]
    │   └── map(1..#, # + #) => [2]
    │       ├── 1..# => [1]
    │       │   ├── 1
    │       │   └── # => 1
    │       └── # + # => [...]
    │           └── (0) ... => 2
    │               └── # + # => 2
    │                   ├── # => 1
    │                   └── # => 1
    ├── (1) ... => [2,4]
    │   └── map(1..#, # + #) => [2,4]
    │       ├── 1..# => [1,2]
    │       │   ├── 1
    │       │   └── # => 2
    │       └── # + # => [...]
    │           ├── (0) ... => 2
    │           │   └── # + # => 2
    │           │       ├── # => 1
    │           │       └── # => 1
    │           └── (1) ... => 4
    │               └── # + # => 4
    │                   ├── # => 2
    │                   └── # => 2
    └── (2) ... => [2,4,6]
        └── map(1..#, # + #) => [2,4,6]
            ├── 1..# => [1,2,3]
            │   ├── 1
            │   └── # => 3
            └── # + # => [...]
                ├── (0) ... => 2
                │   └── # + # => 2
                │       ├── # => 1
                │       └── # => 1
                ├── (1) ... => 4
                │   └── # + # => 4
                │       ├── # => 2
                │       └── # => 2
                └── (2) ... => 6
                    └── # + # => 6
                        ├── # => 3
                        └── # => 3
`,
		},
		{
			"map([1, 2, 3], { let x = #; let y = #*2; map(1..x, { let z = #; x+y+z }) })",
			exprtrace.EvalEnv{},
			`map([1, 2, 3], { let x = #; let y = #*2; map(1..x, { let z = #; x+y+z }) })
│
├── [1, 2, 3] => [1,2,3]
└── let x = #; let y = # * 2; map(1..x, let z = #; x + y + z) => [...]
    ├── (0) ... => [4]
    │   ├── let x = # => 1
    │   ├── let y = # * 2 => 2
    │   │   ├── # => 1
    │   │   └── 2
    │   └── map(1..x, let z = #; x + y + z) => [4]
    │       ├── 1..x => [1]
    │       │   ├── 1
    │       │   └── x => 1
    │       └── let z = #; x + y + z => [...]
    │           └── (0) ... => 4
    │               ├── let z = # => 1
    │               └── x + y + z => 4
    │                   ├── x + y => 3
    │                   │   ├── x => 1
    │                   │   └── y => 2
    │                   └── z => 1
    ├── (1) ... => [7,8]
    │   ├── let x = # => 2
    │   ├── let y = # * 2 => 4
    │   │   ├── # => 2
    │   │   └── 2
    │   └── map(1..x, let z = #; x + y + z) => [7,8]
    │       ├── 1..x => [1,2]
    │       │   ├── 1
    │       │   └── x => 2
    │       └── let z = #; x + y + z => [...]
    │           ├── (0) ... => 7
    │           │   ├── let z = # => 1
    │           │   └── x + y + z => 7
    │           │       ├── x + y => 6
    │           │       │   ├── x => 2
    │           │       │   └── y => 4
    │           │       └── z => 1
    │           └── (1) ... => 8
    │               ├── let z = # => 2
    │               └── x + y + z => 8
    │                   ├── x + y => 6
    │                   │   ├── x => 2
    │                   │   └── y => 4
    │                   └── z => 2
    └── (2) ... => [10,11,12]
        ├── let x = # => 3
        ├── let y = # * 2 => 6
        │   ├── # => 3
        │   └── 2
        └── map(1..x, let z = #; x + y + z) => [10,11,12]
            ├── 1..x => [1,2,3]
            │   ├── 1
            │   └── x => 3
            └── let z = #; x + y + z => [...]
                ├── (0) ... => 10
                │   ├── let z = # => 1
                │   └── x + y + z => 10
                │       ├── x + y => 9
                │       │   ├── x => 3
                │       │   └── y => 6
                │       └── z => 1
                ├── (1) ... => 11
                │   ├── let z = # => 2
                │   └── x + y + z => 11
                │       ├── x + y => 9
                │       │   ├── x => 3
                │       │   └── y => 6
                │       └── z => 2
                └── (2) ... => 12
                    ├── let z = # => 3
                    └── x + y + z => 12
                        ├── x + y => 9
                        │   ├── x => 3
                        │   └── y => 6
                        └── z => 3
`,
		},
		{
			"vars.expected in map(vars.res, { #.value })",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"expected": 10,
					"res":      []any{map[string]any{"value": 10}, map[string]any{"value": 20}, map[string]any{"value": 30}},
				},
			},
			`vars.expected in map(vars.res, { #.value })
│
├── vars.expected => 10
└── map(vars.res, .value) => [10,20,30]
    ├── vars.res => [{"value":10},{"value":20},{"value":30}]
    └── .value => [...]
        ├── (0) ... => 10
        │   └── .value => 10
        ├── (1) ... => 20
        │   └── .value => 20
        └── (2) ... => 30
            └── .value => 30
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			ret, err := EvalWithTrace(tt.in, tt.env)
			if err != nil {
				t.Error(err)
				return
			}
			got, err := ret.FormatTraceTree()
			if err != nil {
				t.Error(err)
				return
			}
			if got != tt.want {
				t.Errorf("got\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func TestEvalWithTraceWithComment(t *testing.T) {
	tests := []struct {
		name string
		in   string
		env  exprtrace.EvalEnv
		want string
	}{
		{
			"No comment",
			"vars.key == 'hello'",
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`vars.key == 'hello'
│
├── vars.key => "hello"
└── "hello"
`,
		},
		{
			"With comment",
			`// This is a comment
vars.key == 'hello' // This is another comment`,
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`// This is a comment
vars.key == 'hello' // This is another comment
│
├── vars.key => "hello"
└── "hello"
`,
		},
		{
			"Deprecated comment annotation",
			`# This is a comment
vars.key == 'hello' # This is another comment`,
			exprtrace.EvalEnv{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`vars.key == 'hello'
│
├── vars.key => "hello"
└── "hello"
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			ret, err := EvalWithTrace(tt.in, tt.env)
			if err != nil {
				t.Error(err)
				return
			}
			got, err := ret.FormatTraceTree()
			if err != nil {
				t.Error(err)
				return
			}
			if got != tt.want {
				t.Errorf("got\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}

func TestEvalCond(t *testing.T) {
	tests := []struct {
		cond  string
		store map[string]any
		want  bool
	}{
		{"hello", map[string]any{
			"hello": true,
		}, true},
		{"hello == 3", map[string]any{
			"hello": 3,
		}, true},
		{"hello == 3", map[string]any{
			"hello": 4,
		}, false},
		{"hello", map[string]any{
			"hello": "true",
		}, false},
		{"hello", nil, false},
	}
	for _, tt := range tests {
		got, err := EvalCond(tt.cond, tt.store)
		if err != nil {
			t.Error(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestTrimComment(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{
			`current.res.status == 200
&& current.res.body.foo == vars.expectFoo
&& current.res.body.bar == vars.expectBar`,
			`current.res.status == 200
&& current.res.body.foo == vars.expectFoo
&& current.res.body.bar == vars.expectBar`,
		},
		{
			`current.res.status == 200
# This is comment
#This is comment
&& current.res.body.foo == vars.expectFoo
#  This is comment
 # This is comment
&& current.res.body.bar == vars.expectBar`,
			`current.res.status == 200
&& current.res.body.foo == vars.expectFoo
&& current.res.body.bar == vars.expectBar`,
		},
		{
			`&& current.res.status == 200 # This is comment.`,
			`&& current.res.status == 200`,
		},
		{
			`&& current.res.status == 200 #This is comment.`,
			`&& current.res.status == 200`,
		},
		{
			`&& current.res.status == 200 # current.res.status == 200`,
			`&& current.res.status == 200`,
		},
		{
			`&& current.res.body.foo == 'Hello # World' # This is comment.`,
			`&& current.res.body.foo == 'Hello # World'`,
		},
		{
			`&& len(map(0..9, {# / 2})) == 5 # This is comment.`,
			`&& len(map(0..9, {# / 2})) == 5`,
		},
		{
			`&& len(map(0..9, {# / 2})) == 5 # len(map(0..9, {# / 2})) == 5 This is comment.`,
			`&& len(map(0..9, {# / 2})) == 5`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got := trimComment(tt.in)
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestEvalCount(t *testing.T) {
	tests := []struct {
		count   string
		store   exprtrace.EvalEnv
		want    int
		wantErr bool
	}{
		{
			"var.count",
			map[string]any{
				"var": map[string]any{
					"count": 3,
				},
			},
			3,
			false,
		},
		{
			"var.count",
			map[string]any{
				"var": map[string]any{
					"count": uint64(3),
				},
			},
			3,
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.count, func(t *testing.T) {
			got, err := EvalCount(tt.count, tt.store)
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvalExpand(t *testing.T) {
	tests := []struct {
		in    any
		store map[string]any
		want  any
	}{
		{
			"",
			map[string]any{},
			"",
		},
		{
			"{{ var }}",
			map[string]any{
				"var": 123,
			},
			uint64(123),
		},
		{
			"{{ var }}",
			map[string]any{
				"var": "123",
			},
			"123",
		},
		{
			"4{{ var }}",
			map[string]any{
				"var": 123,
			},
			uint64(4123),
		},
		{
			"{{ var }}4",
			map[string]any{
				"var": 123,
			},
			uint64(1234),
		},
		{
			"{{ var }}",
			map[string]any{
				"var": false,
			},
			false,
		},
		{
			"{{ var }}",
			map[string]any{
				"var": "false",
			},
			"false",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got, err := EvalExpand(tt.in, tt.store)
			if err != nil {
				t.Error(err)
				return
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}
