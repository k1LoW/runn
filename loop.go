package runn

import (
	"context"
	"strconv"
	"strings"
	"time"

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

	interval    *time.Duration
	minInterval *time.Duration
	maxInterval *time.Duration
}

func newLoop(v interface{}) (*Loop, error) {
	b, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	l := &Loop{}
	err = yaml.Unmarshal(b, l)
	if err != nil {
		// short syntax
		l.Count = strings.TrimRight(string(b), "\n\r")
	}
	if l.Count == "" {
		l.Count = strconv.Itoa(defaultCount)
	}
	if l.Until == "" && l.Interval == "" && l.MinInterval == "" && l.MaxInterval == "" {
		// for simple loop
		l.Interval = "0"
	}
	if l.Until == "" && l.Jitter == nil {
		// for simple loop
		i := 0.0
		l.Jitter = &i
	}
	if l.Interval == "" {
		if l.MinInterval == "" {
			l.MinInterval = defaultMinInterval
		}
		if l.MaxInterval == "" {
			l.MaxInterval = defaultMaxInterval
		}
		if l.Multiplier == nil {
			l.Multiplier = &defaultMultiplier
		}
	}
	if l.Jitter == nil {
		l.Jitter = &defaultJitter
	}

	if l.Interval != "" {
		i, err := parseDuration(l.Interval)
		if err != nil {
			return nil, err
		}
		l.interval = &i
	} else {
		imin, err := parseDuration(l.MinInterval)
		if err != nil {
			return nil, err
		}
		l.minInterval = &imin
		imax, err := parseDuration(l.MaxInterval)
		if err != nil {
			return nil, err
		}
		l.maxInterval = &imax
	}

	return l, nil
}

func (l *Loop) Loop(ctx context.Context) bool {
	if l.ctrl == nil {
		var p backoff.Policy
		if l.interval != nil {
			p = backoff.Constant(
				backoff.WithMaxRetries(0),
				backoff.WithInterval(*l.interval),
				backoff.WithJitterFactor(*l.Jitter),
			)
		} else {
			p = backoff.Exponential(
				backoff.WithMaxRetries(0),
				backoff.WithMinInterval(*l.minInterval),
				backoff.WithMaxInterval(*l.maxInterval),
				backoff.WithMultiplier(*l.Multiplier),
				backoff.WithJitterFactor(*l.Jitter),
			)
		}
		l.ctrl = p.Start(ctx)
	}
	return backoff.Continue(l.ctrl)
}
