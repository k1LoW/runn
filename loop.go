package runn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/k1LoW/duration"
	"github.com/lestrrat-go/backoff/v2"
)

const (
	loopSectionKey            = "loop"
	deprecatedRetrySectionKey = "retry" // deprecated
)

var (
	defaultCount       = 3
	defaultMaxInterval = float64(time.Minute)
	defaultMinInterval = float64(500 * time.Millisecond)
	defaultJitter      = float64(0.0)
	defaultMultiplier  = float64(1.5)
)

type Loop struct {
	Count       *int     `yaml:"count,omitempty"`
	Interval    *float64 `yaml:"interval,omitempty"`
	MinInterval *float64 `yaml:"minInterval,omitempty"`
	MaxInterval *float64 `yaml:"maxInterval,omitempty"`
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
	if err := yaml.Unmarshal(b, r); err != nil {
		return nil, err
	}
	if r.Until == "" {
		return nil, errors.New("until: is empty")
	}
	if r.Count == nil {
		r.Count = &defaultCount
	}
	if r.Jitter == nil {
		r.Jitter = &defaultJitter
	}
	if r.Interval == nil {
		if r.MinInterval == nil {
			r.MinInterval = &defaultMinInterval
		}
		if r.MaxInterval == nil {
			r.MaxInterval = &defaultMaxInterval
		}
		if r.Multiplier == nil {
			r.Multiplier = &defaultMultiplier
		}
	}
	return r, nil
}

func (r *Loop) Loop(ctx context.Context) bool {
	if r.ctrl == nil {
		var p backoff.Policy
		if r.Interval != nil {
			ii, _ := duration.Parse(fmt.Sprintf("%vsec", *r.Interval))
			p = backoff.Constant(
				backoff.WithMaxRetries(*r.Count),
				backoff.WithInterval(ii),
				backoff.WithJitterFactor(*r.Jitter),
			)
		} else {
			imin, _ := duration.Parse(fmt.Sprintf("%vsec", *r.MinInterval))
			imax, _ := duration.Parse(fmt.Sprintf("%vsec", *r.MaxInterval))
			p = backoff.Exponential(
				backoff.WithMaxRetries(*r.Count),
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
