package runn

import (
	"fmt"
	"testing"
)

func TestTrailRunbookID(t *testing.T) {
	tests := []struct {
		trails Trails
		want   string
	}{
		{
			Trails{
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-a",
				},
			},
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
					StepIndex: new(2),
					StepKey:   "s-b",
				},
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-a",
				},
			},
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
					StepIndex: new(3),
					StepKey:   "s-d",
				},
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-c",
				},
				Trail{
					Type:      TrailTypeStep,
					StepIndex: new(2),
					StepKey:   "s-b",
				},
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-a",
				},
			},
			"o-e?step=3&step=2",
		},
		{
			Trails{
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-e",
				},
				Trail{
					Type:      TrailTypeLoop,
					LoopIndex: new(1),
					RunbookID: "o-e",
				},
				Trail{
					Type:      TrailTypeStep,
					StepIndex: new(3),
					StepKey:   "s-d",
				},
				Trail{
					Type:      TrailTypeLoop,
					LoopIndex: new(4),
					StepIndex: new(3),
					StepKey:   "s-d",
				},
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-c",
				},
				Trail{
					Type:      TrailTypeStep,
					StepIndex: new(2),
					StepKey:   "s-b",
				},
				Trail{
					Type:      TrailTypeRunbook,
					RunbookID: "o-a",
				},
			},
			"o-e?step=3&step=2",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := tt.trails.runbookID()
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}
