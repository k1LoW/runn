package runn

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/k1LoW/duration"
	"github.com/lestrrat-go/backoff/v2"
)

const (
	loopSectionKey            = "loop"
	deprecatedRetrySectionKey = "retry" // deprecated
	loopCountVarKey           = "i"
)

var (
	defaultCount       = 3
	defaultMaxInterval = float64(time.Minute)
	defaultMinInterval = float64(500 * time.Millisecond)
	defaultJitter      = float64(0.0)
	defaultMultiplier  = float64(1.5)
)

type Loop struct {
	Count       string   `yaml:"count,omitempty"`
	Interval    *float64 `yaml:"interval,omitempty"`
	MinInterval *float64 `yaml:"minInterval,omitempty"`
	MaxInterval *float64 `yaml:"maxInterval,omitempty"`
	Jitter      *float64 `yaml:"jitter,omitempty"`
	Multiplier  *float64 `yaml:"multiplier,omitempty"`
	Until       string   `yaml:"until"`
	ctrl        backoff.Controller
}

func newLoop(v interface{}) (*Loop, error) {
	b, err := yamlMarshal(v)
	if err != nil {
		return nil, err
	}
	r := &Loop{}
	err = yamlUnmarshal(b, r)
	if err != nil {
		// short syntax
		r.Count = strings.TrimRight(string(b), "\n\r")
	}
	if r.Count == "" {
		r.Count = strconv.Itoa(defaultCount)
	}
	if r.Until == "" && r.Interval == nil && r.MinInterval == nil && r.MaxInterval == nil {
		// for simple loop
		i := 0.0
		r.Interval = &i
	}
	if r.Until == "" && r.Jitter == nil {
		// for simple loop
		i := 0.0
		r.Jitter = &i
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
	if r.Jitter == nil {
		r.Jitter = &defaultJitter
	}
	return r, nil
}

func (r *Loop) Loop(ctx context.Context) bool {
	if r.ctrl == nil {
		var p backoff.Policy
		if r.Interval != nil {
			ii, _ := duration.Parse(fmt.Sprintf("%vsec", *r.Interval))
			p = backoff.Constant(
				backoff.WithMaxRetries(0),
				backoff.WithInterval(ii),
				backoff.WithJitterFactor(*r.Jitter),
			)
		} else {
			imin, _ := duration.Parse(fmt.Sprintf("%vsec", *r.MinInterval))
			imax, _ := duration.Parse(fmt.Sprintf("%vsec", *r.MaxInterval))
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
