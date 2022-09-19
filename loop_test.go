package runn

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNewLoop(t *testing.T) {
	tests := []struct {
		v           interface{}
		count       string
		interval    float64
		minInterval float64
		maxInterval float64
		jitter      float64
		multiplier  float64
		until       string
	}{
		{
			map[string]any{},
			"3",
			0.0,
			*new(float64),
			*new(float64),
			0.0,
			*new(float64),
			"",
		},
		{
			map[string]any{"count": "1", "interval": 1.0, "until": "5", "minInterval": 0.1, "maxInterval": 0.2, "jitter": 0.3, "multiplier": 0.4},
			"1",
			1.0,
			0.1,
			0.2,
			0.3,
			0.4,
			"5",
		},
		{
			map[string]any{"count": "2", "until": "3"},
			"2",
			*new(float64),
			float64(500 * time.Millisecond),
			float64(time.Minute),
			float64(0.0),
			float64(1.5),
			"3",
		},
	}
	for _, tt := range tests {
		got, _ := newLoop(tt.v)
		if diff := cmp.Diff(got.Count, tt.count, nil); diff != "" {
			t.Errorf("Count: %s", diff)
		}
		if got.Interval != nil {
			if diff := cmp.Diff(*got.Interval, tt.interval, nil); diff != "" {
				t.Errorf("Interval: %s", diff)
			}
		}
		if got.MinInterval != nil {
			if diff := cmp.Diff(*got.MinInterval, tt.minInterval, nil); diff != "" {
				t.Errorf("MinInterval: %s", diff)
			}
		}
		if got.MaxInterval != nil {
			if diff := cmp.Diff(*got.MaxInterval, tt.maxInterval, nil); diff != "" {
				t.Errorf("MaxInterval: %s", diff)
			}
		}
		if diff := cmp.Diff(*got.Jitter, tt.jitter, nil); diff != "" {
			t.Errorf("Jitter: %s", diff)
		}
		if got.Multiplier != nil {
			if diff := cmp.Diff(*got.Multiplier, tt.multiplier, nil); diff != "" {
				t.Errorf("Multiplier: %s", diff)
			}
		}
		if diff := cmp.Diff(got.Until, tt.until, nil); diff != "" {
			t.Errorf("Until: %s", diff)
		}
	}
}
