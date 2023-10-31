package runn

import (
	"fmt"
	"testing"
)

func TestTrailRunbookID(t *testing.T) {
	s2 := 2
	s3 := 3

	tests := []struct {
		trails   Trails
		want     string
		wantFull string
	}{
		{
			Trails{
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-a",
				},
			},
			"o-a",
			"o-a",
		},
		{
			Trails{
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-c",
				},
				Trail{
					Type:      TrailTypeStep,
					StepIndex: &s2,
					StepKey:   "s-b",
				},
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-a",
				},
			},
			"o-c",
			"o-c?step=2",
		},
		{
			Trails{
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-e",
				},
				Trail{
					Type:      TrailTypeStep,
					StepIndex: &s3,
					StepKey:   "s-d",
				},
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-c",
				},
				Trail{
					Type:      TrailTypeStep,
					StepIndex: &s2,
					StepKey:   "s-b",
				},
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-a",
				},
			},
			"o-e",
			"o-e?step=3&step=2",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.trails.runbookID()
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
			{
				got := tt.trails.runbookIDFull()
				if got != tt.wantFull {
					t.Errorf("got %v\nwant %v", got, tt.wantFull)
				}
			}
		})
	}
}
