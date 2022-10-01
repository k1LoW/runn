package builtin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		x    interface{}
		y    interface{}
		want bool
	}{
		{1, 1, true},
		{1, 2, false},
		{1, "1", false},
		{"foo", "foo", true},
		{"foo", "bar", false},
		{map[string]interface{}{"foo": "1", "bar": true}, map[string]interface{}{"foo": "1", "bar": true}, true},
		{map[string]interface{}{"foo": "1", "bar": true}, map[string]interface{}{"foo": "1", "bar": false}, false},
	}
	for _, tt := range tests {
		got := Compare(tt.x, tt.y)
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
