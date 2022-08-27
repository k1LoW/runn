package runn

import (
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
└── 'hello' => "hello"
`},
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
		{`body contains "<h1>hello</hello>"`, []string{"body", `"<h1>hello</hello>"`}},
		{`res.body.data.key contains "xxxxxx"`, []string{"res.body.data.key", `"xxxxxx"`}},
		{`res.headers["Content-Type"] == "application/json"`, []string{`res.headers["Content-Type"]`, `"application/json"`}},
		{`res.headers["Content-Type"][0] == "application/json"`, []string{`res.headers["Content-Type"][0]`, `"application/json"`}},
		{`res.body.data.projects[0].name == "myproject"`, []string{`res.body.data.projects[0].name`, `"myproject"`}},
	}
	for _, tt := range tests {
		got := values(tt.cond)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
