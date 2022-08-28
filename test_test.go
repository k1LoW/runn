package runn

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestBuildTree(t *testing.T) {
	tests := []struct {
		cond  string
		store map[string]interface{}
		want  string
	}{
		{
			"vars.key == 'hello'",
			map[string]interface{}{
				"vars": map[string]interface{}{
					"key": "hello",
				},
			},
			`vars.key == 'hello'
├── vars.key => "hello"
└── "hello" => "hello"
`,
		},
		{
			"printf('%s world', vars.key) == 'hello world'",
			map[string]interface{}{
				"vars": map[string]interface{}{
					"key": "hello",
				},
				"printf": func(format string, a ...interface{}) string {
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
	}
	for _, tt := range tests {
		got := buildTree(tt.cond, tt.store)
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
		{`body contains "<h1>hello</hello>"`, []string{"body", `"<h1>hello</hello>"`}},
		{`res.body.data.key contains "xxxxxx"`, []string{"res.body.data.key", `"xxxxxx"`}},
		{`res.headers["Content-Type"] == "application/json"`, []string{`res.headers["Content-Type"]`, `"application/json"`}},
		{`current.raws[0]`, []string{`current.raws[0]`}},
		{`body[0]["key"].data`, []string{`body[0]["key"].data`}},
		{`res.headers["Content-Type"][0] == "application/json"`, []string{`res.headers["Content-Type"][0]`, `"application/json"`}},
		{`res.body.data.projects[0].name == "myproject"`, []string{`res.body.data.projects[0].name`, `"myproject"`}},
		{`printf('%s world', vars.key) == 'hello world'`, []string{`printf("%s world", vars.key)`, `"%s world"`, `vars.key`, `"hello world"`}},
		{`compare(steps[8].res.body, vars.wantBody, "Content-Length")`, []string{`compare(steps[8].res.body, vars.wantBody, "Content-Length")`, `steps[8].res.body`, `vars.wantBody`, `"Content-Length"`}},
	}
	for _, tt := range tests {
		got := values(tt.cond)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
