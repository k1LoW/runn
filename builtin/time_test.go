package builtin

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		v    any
		want time.Time
	}{
		{now.String(), now},
		{"err", time.Time{}},
	}
	for _, tt := range tests {
		got := Time(tt.v)
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Error(diff)
		}
	}
}
