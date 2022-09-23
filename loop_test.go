package runn

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewLoop(t *testing.T) {
	tests := []struct {
		v           interface{}
		count       string
		interval    string
		minInterval string
		maxInterval string
		jitter      float64
		multiplier  float64
		until       string
	}{
		{
			map[string]any{},
			"3",
			"0",
			"",
			"",
			0.0,
			*new(float64),
			"",
		},
		{
			map[string]any{"count": "1", "interval": "1.0", "until": "5", "minInterval": "0.1", "maxInterval": "0.2", "jitter": 0.3, "multiplier": 0.4},
			"1",
			"1.0",
			"0.1",
			"0.2",
			0.3,
			0.4,
			"5",
		},
		{
			map[string]any{"count": "2", "until": "3"},
			"2",
			"0",
			"500ms",
			"1min",
			float64(0.0),
			float64(1.5),
			"3",
		},
	}
	for _, tt := range tests {
		got, err := newLoop(tt.v)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(got.Count, tt.count, nil); diff != "" {
			t.Errorf("Count: %s", diff)
		}
		if got.Interval != "" {
			if diff := cmp.Diff(got.Interval, tt.interval, nil); diff != "" {
				t.Errorf("Interval: %s", diff)
			}
		}
		if got.MinInterval != "" {
			if diff := cmp.Diff(got.MinInterval, tt.minInterval, nil); diff != "" {
				t.Errorf("MinInterval: %s", diff)
			}
		}
		if got.MaxInterval != "" {
			if diff := cmp.Diff(got.MaxInterval, tt.maxInterval, nil); diff != "" {
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
