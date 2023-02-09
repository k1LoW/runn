package builtin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIntersect(t *testing.T) {
	tests := []struct {
		x    interface{}
		y    interface{}
		want interface{}
	}{
		{[]int{1, 2}, []int{1, 2, 3}, []any{int(1), int(2)}},
		{[]int{1, 2, 3}, []int{1, 2, 3}, []any{int(1), int(2), int(3)}},
		{[]int{1, 2, 3, 4}, []int{1, 2, 3}, []any{int(1), int(2), int(3)}},
		{[]string{"a", "b"}, []string{"b", "c", "d"}, []any{string("b")}},
	}
	for _, tt := range tests {
		got := Intersect(tt.x, tt.y)
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
