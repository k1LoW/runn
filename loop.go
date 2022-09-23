package runn

import (
	"context"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/lestrrat-go/backoff/v2"
)

const (
	loopSectionKey            = "loop"
	deprecatedRetrySectionKey = "retry" // deprecated
	loopCountVarKey           = "i"
)

var (
	defaultCount       = 3
	defaultMaxInterval = "1min"
	defaultMinInterval = "500ms"
	defaultJitter      = float64(0.0)
	defaultMultiplier  = float64(1.5)
)

type Loop struct {
	Count       string   `yaml:"count,omitempty"`
	Interval    string   `yaml:"interval,omitempty"`
	MinInterval string   `yaml:"minInterval,omitempty"`
	MaxInterval string   `yaml:"maxInterval,omitempty"`
	Jitter      *float64 `yaml:"jitter,omitempty"`
	Multiplier  *float64 `yaml:"multiplier,omitempty"`
	Until       string   `yaml:"until"`
	ctrl        backoff.Controller
}

func newLoop(v interface{}) (*Loop, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	r := &Loop{}
	err = yaml.Unmarshal(b, r)
	if err != nil {
		// short syntax
		r.Count = strings.TrimRight(string(b), "\n\r")
	}
	if r.Count == "" {
		r.Count = strconv.Itoa(defaultCount)
	}
	if r.Until == "" && r.Interval == "" && r.MinInterval == "" && r.MaxInterval == "" {
		// for simple loop
		r.Interval = "0"
	}
	if r.Until == "" && r.Jitter == nil {
		// for simple loop
		i := 0.0
		r.Jitter = &i
	}
	if r.Interval == "" {
		if r.MinInterval == "" {
			r.MinInterval = defaultMinInterval
		}
		if r.MaxInterval == "" {
			r.MaxInterval = defaultMaxInterval
		}
		if r.Multiplier == nil {
			r.Multiplier = &defaultMultiplier
		}
	}
	if r.Jitter == nil {
		r.Jitter = &defaultJitter
	}
	return r, nil
}

func (r *Loop) Loop(ctx context.Context) bool {
	if r.ctrl == nil {
		var p backoff.Policy
		if r.Interval != "" {
			ii, _ := parseDuration(r.Interval)
			p = backoff.Constant(
				backoff.WithMaxRetries(0),
				backoff.WithInterval(ii),
				backoff.WithJitterFactor(*r.Jitter),
			)
		} else {
			imin, _ := parseDuration(r.MinInterval)
			imax, _ := parseDuration(r.MaxInterval)
			p = backoff.Exponential(
				backoff.WithMaxRetries(0),
				backoff.WithMinInterval(imin),
				backoff.WithMaxInterval(imax),
				backoff.WithMultiplier(*r.Multiplier),
				backoff.WithJitterFactor(*r.Jitter),
			)
		}
		r.ctrl = p.Start(ctx)
	}
	return backoff.Continue(r.ctrl)
}
