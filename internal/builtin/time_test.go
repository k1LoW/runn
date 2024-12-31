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
		{now.UnixNano(), now},
		{now.String(), now},
		{"err", time.Time{}},
	}
	for _, tt := range tests {
		got, err := Time(tt.v)
		if err != nil {
			zero := time.Time{}
			if tt.want.UnixNano() != zero.UnixNano() {
				t.Error(err)
			}
			continue
		}
		if diff := cmp.Diff(got, tt.want); diff != "" {
			t.Error(diff)
		}
	}
}
