package runn

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
