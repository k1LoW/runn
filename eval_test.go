package runn

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/runn/builtin"
)

func TestBuildTree(t *testing.T) {
	tests := []struct {
		cond  string
		store map[string]any
		want  string
	}{
		{
			"vars.key == 'hello'",
			map[string]any{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`vars.key == 'hello'
├── vars.key => "hello"
└── "hello" => "hello"
`,
		},
		{
			"vars['key'] == 'hello'",
			map[string]any{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`vars['key'] == 'hello'
├── vars.key => "hello"
└── "hello" => "hello"
`,
		},
		{
			`# comment
1 == 1`,
			map[string]any{},
			`1 == 1
├── 1 => 1
└── 1 => 1
`,
		},
		{
			"printf('%s world', vars.key) == 'hello world'",
			map[string]any{
				"vars": map[string]any{
					"key": "hello",
				},
				"printf": func(format string, a ...any) string {
					return fmt.Sprintf(format, a...)
				},
			},
			`printf('%s world', vars.key) == 'hello world'
├── printf("%s world", vars.key) => "hello world"
├── "%s world" => "%s world"
├── vars.key => "hello"
└── "hello world" => "hello world"
`,
		},
		{
			"vars.tests[i] == vars.wants[i]",
			map[string]any{
				"vars": map[string]any{
					"tests": []int{1, 2, 3},
					"wants": []int{0, 1, 2},
				},
				"i": 1,
			},
			`vars.tests[i] == vars.wants[i]
├── vars.tests[i] => 2
└── vars.wants[i] => 1
`,
		},
		{
			"vars.expected in map(vars.res, { #.value })",
			map[string]any{
				"vars": map[string]any{
					"expected": 10,
					"res":      []any{map[string]any{"value": 10}, map[string]any{"value": 20}},
				},
			},
			`vars.expected in map(vars.res, { #.value })
├── vars.expected => 10
├── map(vars.res, { #.value }) => [10,20]
└── vars.res => [{"value":10},{"value":20}]
`,
		},
		{
			"diff(intersect(vars.v1, vars.v2), vars.v3) == ''",
			map[string]any{
				"vars": map[string]any{
					"v1": []any{1, 2, 3, 4},
					"v2": []any{1, 3, 5, 7},
					"v3": []any{1, 3},
				},
				"diff":      builtin.Diff,
				"intersect": builtin.Intersect,
			},
			`diff(intersect(vars.v1, vars.v2), vars.v3) == ''
├── diff(intersect(vars.v1, vars.v2), vars.v3) => ""
├── intersect(vars.v1, vars.v2) => [1,3]
├── vars.v3 => [1,3]
├── vars.v1 => [1,2,3,4]
├── vars.v2 => [1,3,5,7]
└── "" => ""
`,
		},
		{
			`vars.key == "hello\nworld"`,
			map[string]any{
				"vars": map[string]any{
					"key": "hello",
				},
			},
			`vars.key == "hello\nworld"
├── vars.key => "hello"
└── "hello\nworld" => "hello
    world"
`,
		},
	}
	for _, tt := range tests {
		got, err := buildTree(tt.cond, tt.store)
		if err != nil {
			t.Error(err)
		}
		if got != tt.want {
			t.Errorf("got\n%v\nwant\n%v", got, tt.want)
		}
	}
}

func TestValues(t *testing.T) {
	tests := []struct {
		cond string
		want []string
	}{
		{`"Content-Type" in headers`, []string{`"Content-Type"`, "headers"}},
		{`1 + 2`, []string{`1`, `2`}},
		{`1.5 - 4.55`, []string{`1.5`, `4.55`}},
		{`1..3 == [1, 2, 3]`, []string{`1`, `3`, `[1, 2, 3]`}},
		{`"foo" in {foo: 1, bar: 2}`, []string{`"foo"`, `{foo: 1, bar: 2}`}},
		{`true != false`, []string{`true`, `false`}},
		{`nil`, []string{`<nil>`}},
		{`body contains "<h1>hello</hello>"`, []string{"body", `"<h1>hello</hello>"`}},
		{`res.body.data.key contains "xxxxxx"`, []string{"res.body.data.key", `"xxxxxx"`}},
		{`res.headers["Content-Type"] == "application/json"`, []string{`res.headers["Content-Type"]`, `"application/json"`}},
		{`current.rows[0]`, []string{`current.rows[0]`}},
		{`body[0]["key"].data`, []string{`body[0].key.data`}},
		{`res.headers["Content-Type"][0] == "application/json"`, []string{`res.headers["Content-Type"][0]`, `"application/json"`}},
		{`res.body.data.projects[0].name == "myproject"`, []string{`res.body.data.projects[0].name`, `"myproject"`}},
		{`printf('%s world', vars.key) == 'hello world'`, []string{`printf("%s world", vars.key)`, `"%s world"`, `vars.key`, `"hello world"`}},
		{`compare(steps[8].res.body, vars.wantBody, "Content-Length")`, []string{`compare(steps[8].res.body, vars.wantBody, "Content-Length")`, `steps[8].res.body`, `vars.wantBody`, `"Content-Length"`}},
		{`len("hello")`, []string{`len("hello")`, `"hello"`}},
		{`vars.tests[i]`, []string{`vars.tests[i]`}},
		{`vars.tests[i] == vars.wants[j]`, []string{`vars.tests[i]`, `vars.wants[j]`}},
	}
	for _, tt := range tests {
		got, err := values(tt.cond)
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
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
				t.Errorf("%s", diff)
			}
		})
	}
}
