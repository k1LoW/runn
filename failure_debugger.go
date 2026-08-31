package runn

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/k1LoW/maskedio"
)

const failureDebugMaxBytes = 1 << 20

type failureDebugCoordinator struct {
	mu sync.Mutex
}

func newFailureDebugCoordinator() *failureDebugCoordinator {
	return &failureDebugCoordinator{}
}

func (c *failureDebugCoordinator) write(out *maskedio.Writer, b []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := out.Write(b)
	return errors.Join(err, out.Flush())
}

type limitedBuffer struct {
	buf       bytes.Buffer
	max       int
	truncated bool
}

func newLimitedBuffer(maxBytes int) *limitedBuffer {
	return &limitedBuffer{max: maxBytes}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.max - b.buf.Len()
	if remaining <= 0 {
		b.truncated = b.truncated || n > 0
		return n, nil
	}
	if len(p) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.truncated = true
		return n, nil
	}
	_, _ = b.buf.Write(p)
	return n, nil
}

func (b *limitedBuffer) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *limitedBuffer) Len() int {
	return b.buf.Len()
}

func (b *limitedBuffer) Reset() {
	b.buf.Reset()
	b.truncated = false
}

type failureDebugger struct {
	*debugger
	buf         *limitedBuffer
	out         *maskedio.Writer
	coordinator *failureDebugCoordinator
}

func newFailureDebugger(out *maskedio.Writer, coordinator *failureDebugCoordinator) *failureDebugger {
	buf := newLimitedBuffer(failureDebugMaxBytes)
	return &failureDebugger{
		debugger:    NewDebugger(buf),
		buf:         buf,
		out:         out,
		coordinator: coordinator,
	}
}

func (d *failureDebugger) CaptureResultByStep(_ Trails, result *RunResult) {
	defer d.buf.Reset()

	sr := latestStepResult(result.StepResults)
	if sr == nil || sr.Err == nil || d.buf.Len() == 0 {
		return
	}

	var out bytes.Buffer
	_, _ = fmt.Fprintln(&out, "-----START FAILURE DIAGNOSTICS-----")
	_, _ = fmt.Fprintf(&out, "Runbook: %s\n", normalizePath(result.Path))
	_, _ = fmt.Fprintf(&out, "Step: %s\n", sr.Key)
	if sr.RunnerType != "" {
		_, _ = fmt.Fprintf(&out, "Runner: %s", sr.RunnerType)
		if sr.RunnerKey != "" {
			_, _ = fmt.Fprintf(&out, " (%s)", sr.RunnerKey)
		}
		_, _ = fmt.Fprintln(&out)
	}
	_, _ = fmt.Fprintln(&out)
	_, _ = out.Write(d.buf.Bytes())
	if d.buf.truncated {
		_, _ = fmt.Fprintf(&out, "\n... diagnostics truncated at %d bytes ...\n", failureDebugMaxBytes)
	}
	_, _ = fmt.Fprintln(&out, "-----END FAILURE DIAGNOSTICS-----")

	if err := d.coordinator.write(d.out, out.Bytes()); err != nil {
		d.errs = errors.Join(d.errs, err)
	}
}

func (d *failureDebugger) SetCurrentTrails(trs Trails) {
	d.buf.Reset()
	d.debugger.SetCurrentTrails(trs)
}

func latestStepResult(results []*StepResult) *StepResult {
	for _, result := range slices.Backward(results) {
		if result != nil {
			return result
		}
	}
	return nil
}
