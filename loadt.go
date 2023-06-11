package runn

import (
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	or "github.com/ryo-yamaoka/otchkiss/result"
)

const reportTemplate = `
Number of runbooks per RunN...: {{ .NumberOfRunbooks }}
Warm up time (--warm-up)......: {{ .WarmUpTime }}
Duration (--duration).........: {{ .Duration }}
Concurrent (--load-concurrent): {{ .MaxConcurrent }}

Total.........................: {{ .TotalRequests }}
Succeeded.....................: {{ .Succeeded }}
Failed........................: {{ .Failed }}
Error rate....................: {{ .ErrorRate }}%
RunN per seconds..............: {{ .RPS }}
Latency ......................: max={{ .MaxLatency }}ms min={{ .MinLatency }}ms avg={{ .AvgLatency }}ms med={{ .MedLatency }}ms p(90)={{ .Latency90p }}ms p(99)={{ .Latency99p }}ms
`

type loadtResult struct {
	runbookCount int64
	warmUp       time.Duration
	duration     time.Duration
	concurrent   int64
	total        int64
	succeeded    int64
	failed       int64
	errorRate    float64
	rps          float64
	max          float64
	min          float64
	p99          float64
	p90          float64
	p50          float64
	avg          float64
}

func NewLoadtResult(rc int, w, d time.Duration, c int, r *or.Result) (*loadtResult, error) {
	succeeded := r.Succeeded()
	failed := r.Failed()
	total := succeeded + failed
	er := float64(failed) / float64(total) * 100
	rps := float64(total) / d.Seconds()

	max, err := r.PercentileLatency(100)
	if err != nil {
		return nil, err
	}
	min, err := r.PercentileLatency(0)
	if err != nil {
		return nil, err
	}
	p99, err := r.PercentileLatency(99)
	if err != nil {
		return nil, err
	}
	p90, err := r.PercentileLatency(90)
	if err != nil {
		return nil, err
	}
	p50, err := r.PercentileLatency(50)
	if err != nil {
		return nil, err
	}

	ll := r.Latencies()
	var avg float64
	for _, l := range ll {
		avg += l
	}
	avg = avg / float64(len(ll))

	return &loadtResult{
		runbookCount: int64(rc),
		warmUp:       w,
		duration:     d,
		concurrent:   int64(c),
		total:        total,
		succeeded:    succeeded,
		failed:       failed,
		errorRate:    er,
		rps:          rps,
		max:          max,
		min:          min,
		p99:          p99,
		p90:          p90,
		p50:          p50,
		avg:          avg,
	}, nil
}

func (r *loadtResult) Report(w io.Writer) error {
	tmpl, err := template.New("report").Parse(reportTemplate)
	if err != nil {
		return err
	}
	data := map[string]any{
		"NumberOfRunbooks": r.runbookCount,
		"WarmUpTime":       r.warmUp.String(),
		"Duration":         r.duration.String(),
		"MaxConcurrent":    r.concurrent,
		"TotalRequests":    r.total,
		"Succeeded":        r.succeeded,
		"Failed":           r.failed,
		"ErrorRate":        humanize.CommafWithDigits(r.errorRate, 1),
		"RPS":              humanize.CommafWithDigits(r.rps, 1),
		"MaxLatency":       humanize.CommafWithDigits(r.max, 1),
		"MinLatency":       humanize.CommafWithDigits(r.min, 1),
		"AvgLatency":       humanize.CommafWithDigits(r.avg, 1),
		"MedLatency":       humanize.CommafWithDigits(r.p50, 1),
		"Latency90p":       humanize.CommafWithDigits(r.p90, 1),
		"Latency99p":       humanize.CommafWithDigits(r.p99, 1),
	}
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}
	return nil
}

func (r *loadtResult) CheckThreshold(threshold string) error {
	if threshold == "" {
		return nil
	}
	store := map[string]any{
		"total":      r.total,
		"succeeded":  r.succeeded,
		"failed":     r.failed,
		"error_rate": r.errorRate,
		"rps":        r.rps,
		"max":        r.max * 1000,
		"mid":        r.p50 * 1000,
		"min":        r.min * 1000,
		"p90":        r.p90 * 1000,
		"p99":        r.p99 * 1000,
		"avg":        r.avg * 1000,
	}
	tf, err := EvalCond(threshold, store)
	if err != nil {
		return err
	}
	if !tf {
		bt, err := buildTree(threshold, store)
		if err != nil {
			return err
		}
		return fmt.Errorf("(%s) is not true\n%s", threshold, bt)
	}
	return nil
}
