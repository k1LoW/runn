package runn

import (
	"testing"
)

func TestBuildTree(t *testing.T) {
	tests := []struct {
		cond  string
		store map[string]interface{}
		want  string
	}{
		{"vars.key == 'hello'", map[string]interface{}{
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
