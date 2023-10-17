package runn

import (
	"fmt"
	"testing"
)

func TestStepRunbookID(t *testing.T) {
	s2 := 2
	s3 := 3

	tests := []struct {
		s        *step
		want     string
		wantFull string
	}{
		{
			&step{idx: s2, key: "s-b", parent: &operator{id: "o-c"}},
			"o-c",
			"o-c?step=2",
		},
		{
			&step{idx: s2, key: "s-b", parent: &operator{id: "o-c", parent: &step{idx: s3, key: "s-d", parent: &operator{id: "o-e"}}}},
			"o-e",
			"o-e?step=3&step=2",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.s.runbookID()
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
			{
				got := tt.s.runbookIDFull()
				if got != tt.wantFull {
					t.Errorf("got %v\nwant %v", got, tt.wantFull)
				}
			}
		})
	}
}
