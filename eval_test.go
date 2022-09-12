package runn

import (
	"testing"
)

func TestEvalCond(t *testing.T) {
	tests := []struct {
		cond  string
		store map[string]interface{}
		want  bool
	}{
		{"hello", map[string]interface{}{
			"hello": true,
		}, true},
		{"hello == 3", map[string]interface{}{
			"hello": 3,
		}, true},
		{"hello == 3", map[string]interface{}{

			"hello": 4,
		}, false},
		{"hello", map[string]interface{}{
			"hello": "true",
		}, false},
		{"hello", nil, false},
	}
	for _, tt := range tests {
		got, err := evalCond(tt.cond, tt.store)
		if err != nil {
			t.Error(err)
		}
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}
