package runn

import (
	"encoding/json"
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/k1LoW/runn/internal/expr"
	or "github.com/ryo-yamaoka/otchkiss/result"
)

type loadtResultJSON struct {
	RunbookCount int64   `json:"runbook_count"`
	WarmUp       string  `json:"warm_up"`
	Duration     string  `json:"duration"`
	Concurrent   int64   `json:"concurrent"`
	MaxRPS       int64   `json:"max_rps"`
	Total        int64   `json:"total"`
	Succeeded    int64   `json:"succeeded"`
	Failed       int64   `json:"failed"`
	ErrorRate    float64 `json:"error_rate"`
	RPS          float64 `json:"rps"`
	LatencyMaxMs float64 `json:"latency_max_ms"`
	LatencyMinMs float64 `json:"latency_min_ms"`
	LatencyAvgMs float64 `json:"latency_avg_ms"`
	LatencyMedMs float64 `json:"latency_med_ms"`
	LatencyP90Ms float64 `json:"latency_p90_ms"`
	LatencyP99Ms float64 `json:"latency_p99_ms"`
}

const reportTemplate = `
Number of runbooks per RunN....: {{ .NumberOfRunbooks }}
Warm up time (--warm-up).......: {{ .WarmUpTime }}
Duration (--duration)..........: {{ .Duration }}
Concurrent (--load-concurrent).: {{ .MaxConcurrent }}
Max RunN per second (--max-rps): {{ .MaxRPS }}

Total..........................: {{ .TotalRequests }}
Succeeded......................: {{ .Succeeded }}
Failed.........................: {{ .Failed }}
Error rate.....................: {{ .ErrorRate }}%
RunN per second................: {{ .RPS }}
Latency .......................: max={{ .MaxLatency }}ms min={{ .MinLatency }}ms avg={{ .AvgLatency }}ms med={{ .MedLatency }}ms p(90)={{ .Latency90p }}ms p(95)={{ .Latency95p }}ms p(99)={{ .Latency99p }}ms

`

type loadtResult struct {
	runbookCount int64
	warmUp       time.Duration
	duration     time.Duration
	concurrent   int64
	maxRPS       int64
	total        int64
	succeeded    int64
	failed       int64
	errorRate    float64
	rps          float64
	max          float64
	min          float64
	p99          float64
	p95          float64
	p90          float64
	p50          float64
	avg          float64
}

// NewLoadtResult creates a new load test result with the provided parameters.
// It calculates various metrics such as error rate, requests per second, and percentiles.
func NewLoadtResult(rc int, w, d time.Duration, c, m int, r *or.Result) (*loadtResult, error) {
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
	p95, err := r.PercentileLatency(95)
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
		maxRPS:       int64(m),
		total:        total,
		succeeded:    succeeded,
		failed:       failed,
		errorRate:    er,
		rps:          rps,
		max:          max,
		min:          min,
		p99:          p99,
		p95:          p95,
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
		"MaxRPS":           r.maxRPS,
		"TotalRequests":    r.total,
		"Succeeded":        r.succeeded,
		"Failed":           r.failed,
		"ErrorRate":        humanize.CommafWithDigits(r.errorRate, 1),
		"RPS":              humanize.CommafWithDigits(r.rps, 1),
		"MaxLatency":       humanize.CommafWithDigits(r.max*1000, 1),
		"MinLatency":       humanize.CommafWithDigits(r.min*1000, 1),
		"AvgLatency":       humanize.CommafWithDigits(r.avg*1000, 1),
		"MedLatency":       humanize.CommafWithDigits(r.p50*1000, 1),
		"Latency90p":       humanize.CommafWithDigits(r.p90*1000, 1),
		"Latency95p":       humanize.CommafWithDigits(r.p95*1000, 1),
		"Latency99p":       humanize.CommafWithDigits(r.p99*1000, 1),
	}
	if err := tmpl.Execute(w, data); err != nil {
		return err
	}
	return nil
}

// ReportJSON writes the load test result as JSON.
func (r *loadtResult) ReportJSON(w io.Writer) error {
	j := loadtResultJSON{
		RunbookCount: r.runbookCount,
		WarmUp:       r.warmUp.String(),
		Duration:     r.duration.String(),
		Concurrent:   r.concurrent,
		MaxRPS:       r.maxRPS,
		Total:        r.total,
		Succeeded:    r.succeeded,
		Failed:       r.failed,
		ErrorRate:    r.errorRate,
		RPS:          r.rps,
		LatencyMaxMs: r.max * 1000,
		LatencyMinMs: r.min * 1000,
		LatencyAvgMs: r.avg * 1000,
		LatencyMedMs: r.p50 * 1000,
		LatencyP90Ms: r.p90 * 1000,
		LatencyP99Ms: r.p99 * 1000,
	}
	b, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, string(b)); err != nil {
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
		"p95":        r.p95 * 1000,
		"p99":        r.p99 * 1000,
		"avg":        r.avg * 1000,
	}
	tf, err := expr.EvalWithTrace(threshold, store)
	if err != nil {
		return err
	}
	if !tf.OutputAsBool() {
		bt, err := tf.FormatTraceTree()
		if err != nil {
			return err
		}
		return fmt.Errorf("(%s) is not true\n%s", threshold, bt)
	}
	return nil
}
