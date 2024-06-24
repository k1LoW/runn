package capture

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/k1LoW/runn"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
)

var _ runn.Capturer = (*cRunbook)(nil)

type cRunbook struct {
	dir           string
	currentTrails runn.Trails
	errs          error
	runbooks      sync.Map
	loadDesc      bool
	desc          string
	labels        []string
	runners       map[string]any
}

type runbookMeta struct {
	Desc    string        `yaml:"desc"`
	Labels  []string      `yaml:"labels,omitempty"`
	Runners yaml.MapSlice `yaml:"runners,omitempty"`
}

type runbook struct {
	Desc    string          `yaml:"desc"`
	Labels  []string        `yaml:"labels,omitempty"`
	Runners yaml.MapSlice   `yaml:"runners,omitempty"`
	Steps   []yaml.MapSlice `yaml:"steps"`

	currentGRPCType          runn.GRPCType
	currentGRPCStatus        *status.Status
	currentGRPCResponceIndex int
	currentGRPCTestCond      []string
	currentExecTestCond      []string
}

type RunbookOption func(*cRunbook) error

func RunbookLoadDesc(enable bool) RunbookOption {
	return func(r *cRunbook) error {
		r.loadDesc = enable
		return nil
	}
}

func Runbook(dir string, opts ...RunbookOption) *cRunbook {
	r := &cRunbook{
		dir:      dir,
		runbooks: sync.Map{},
		runners:  map[string]any{},
	}
	for _, opt := range opts {
		_ = opt(r)
	}
	return r
}

func (c *cRunbook) CaptureStart(trs runn.Trails, bookPath, desc string) {
	if _, err := os.Stat(bookPath); err == nil {
		func() {
			b, err := os.ReadFile(bookPath)
			if err != nil {
				c.errs = errors.Join(c.errs, err)
				return
			}
			rb := runbookMeta{}
			if err := yaml.Unmarshal(b, &rb); err != nil {
				c.errs = errors.Join(c.errs, err)
				return
			}
			if c.loadDesc {
				c.desc = rb.Desc
			}
			c.labels = rb.Labels
			for _, r := range rb.Runners {
				k, ok := r.Key.(string)
				if !ok {
					continue
				}
				v := r.Value
				c.runners[k] = v
			}
		}()
	}

	c.runbooks.Store(trs[0], &runbook{})
}

func (c *cRunbook) CaptureResult(trs runn.Trails, result *runn.RunResult) {
	if !result.Skipped {
		c.writeRunbook(trs, result.Path)
	}
}

func (c *cRunbook) CaptureEnd(trs runn.Trails, bookPath, desc string) {}

func (c *cRunbook) CaptureResultByStep(trs runn.Trails, result *runn.RunResult) {}

func (c *cRunbook) CaptureHTTPRequest(name string, req *http.Request) {
	const dummyDsn = "[THIS IS HTTP RUNNER]"
	if v, ok := c.runners[name]; ok {
		c.setRunner(name, v)
	} else {
		c.setRunner(name, dummyDsn)
	}
	r := c.currentRunbook()
	if r == nil {
		return
	}

	step, err := runn.CreateHTTPStepMapSlice(name, req)
	if err != nil {
		c.errs = errors.Join(c.errs, err)
		return
	}

	r.Steps = append(r.Steps, step)
}

func (c *cRunbook) CaptureHTTPResponse(name string, res *http.Response) {
	r := c.currentRunbook()
	step := r.latestStep()
	// status
	var cond []string
	cond = append(cond, fmt.Sprintf("current.res.status == %d", res.StatusCode))

	// headers
	var keys []string
	for k := range res.Header {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if k == "Date" {
			cond = append(cond, fmt.Sprintf("%q in current.res.headers", k))
			continue
		}
		for i, v := range res.Header[k] {
			cond = append(cond, fmt.Sprintf("current.res.headers[%q][%d] == %#v", k, i, v))
		}
	}

	// body
	contentType := res.Header.Get("Content-Type")
	var (
		save io.ReadCloser
		err  error
	)
	save, res.Body, err = drainBody(res.Body)
	if err != nil {
		c.errs = errors.Join(c.errs, fmt.Errorf("failed to drainBody: %w", err))
		return
	}
	if strings.Contains(contentType, "json") {
		b, err := io.ReadAll(save)
		if err != nil {
			c.errs = errors.Join(c.errs, fmt.Errorf("failed to io.ReadAll: %w", err))
			return
		}
		var v any
		if err := json.Unmarshal(b, &v); err != nil {
			c.errs = errors.Join(c.errs, fmt.Errorf("failed to json.Unmarshal: %w", err))
			return
		}
		b, err = json.Marshal(v)
		if err != nil {
			c.errs = errors.Join(c.errs, fmt.Errorf("failed to json.Marshal: %w", err))
			return
		}
		cond = append(cond, fmt.Sprintf("compare(current.res.body, %s)", string(b)))

	} else {
		b, err := io.ReadAll(save)
		if err != nil {
			c.errs = errors.Join(c.errs, fmt.Errorf("failed to io.ReadAll: %w", err))
			return
		}
		cond = append(cond, fmt.Sprintf("current.res.rawBody == %#v", string(b)))
	}

	r.replaceLatestStep(append(step, yaml.MapItem{Key: "test", Value: fmt.Sprintf("%s\n", strings.Join(cond, "\n&& "))}))
}

func (c *cRunbook) CaptureGRPCStart(name string, typ runn.GRPCType, service, method string) {
	const dummyDsn = "[THIS IS gRPC RUNNER]"
	if v, ok := c.runners[name]; ok {
		c.setRunner(name, v)
	} else {
		c.setRunner(name, dummyDsn)
	}
	r := c.currentRunbook()
	if r == nil {
		return
	}
	r.currentGRPCType = typ
	step := yaml.MapSlice{
		{Key: name, Value: yaml.MapSlice{
			{Key: fmt.Sprintf("%s/%s", service, method), Value: yaml.MapSlice{}},
		}},
	}
	r.Steps = append(r.Steps, step)
}

func (c *cRunbook) CaptureGRPCRequestHeaders(h map[string][]string) {
	if len(h) == 0 {
		return
	}
	hh := map[string]any{}
	var keys []string
	for k := range h {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if len(h[k]) == 1 {
			hh[k] = h[k][0]
			continue
		}
		hh[k] = h[k]
	}
	r := c.currentRunbook()
	step := r.latestStep()
	hb := headersAndMessages(step)
	hb = append(hb, yaml.MapItem{Key: "headers", Value: hh})
	step = replaceHeadersAndMessages(step, hb)
	r.replaceLatestStep(step)
}

func (c *cRunbook) CaptureGRPCRequestMessage(m map[string]any) {
	if len(m) == 0 {
		return
	}
	r := c.currentRunbook()
	step := r.latestStep()
	hb := headersAndMessages(step)
	switch r.currentGRPCType {
	case runn.GRPCUnary, runn.GRPCServerStreaming:
		hb = append(hb, yaml.MapItem{Key: "message", Value: m})
	case runn.GRPCClientStreaming, runn.GRPCBidiStreaming:
		hb = c.appendOp(hb, m)
	}
	step = replaceHeadersAndMessages(step, hb)
	r.replaceLatestStep(step)
}

func (c *cRunbook) CaptureGRPCResponseStatus(s *status.Status) {
	r := c.currentRunbook()
	r.currentGRPCStatus = s
}

func (c *cRunbook) CaptureGRPCResponseHeaders(h map[string][]string) {
	c.captureGRPCResponseMetadata("headers", h)
}

func (c *cRunbook) CaptureGRPCResponseMessage(m map[string]any) {
	r := c.currentRunbook()
	step := r.latestStep()
	hb := headersAndMessages(step)
	switch r.currentGRPCType {
	case runn.GRPCBidiStreaming:
		hb = c.appendOp(hb, runn.GRPCOpReceive)
	}

	b, err := json.Marshal(m)
	if err != nil {
		c.errs = errors.Join(c.errs, fmt.Errorf("failed to yaml.Marshal: %w", err))
		return
	}
	switch r.currentGRPCType {
	case runn.GRPCUnary, runn.GRPCClientStreaming:
		cond := fmt.Sprintf("compare(current.res.message, %s)", string(b))
		r.currentGRPCTestCond = append(r.currentGRPCTestCond, cond)
	case runn.GRPCServerStreaming, runn.GRPCBidiStreaming:
		cond := fmt.Sprintf("compare(current.res.messages[%d], %s)", r.currentGRPCResponceIndex, string(b))
		r.currentGRPCTestCond = append(r.currentGRPCTestCond, cond)
	}

	step = replaceHeadersAndMessages(step, hb)
	r.replaceLatestStep(step)
	r.currentGRPCResponceIndex += 1
}

func (c *cRunbook) CaptureGRPCResponseTrailers(t map[string][]string) {
	c.captureGRPCResponseMetadata("trailers", t)
}

func (c *cRunbook) CaptureGRPCClientClose() {
	r := c.currentRunbook()
	step := r.latestStep()
	hb := headersAndMessages(step)
	switch r.currentGRPCType {
	case runn.GRPCBidiStreaming:
		hb = c.appendOp(hb, runn.GRPCOpClose)
	}
	step = replaceHeadersAndMessages(step, hb)
	r.replaceLatestStep(step)
}

func (c *cRunbook) CaptureGRPCEnd(name string, typ runn.GRPCType, service, method string) {
	r := c.currentRunbook()
	cond := fmt.Sprintf("current.res.status == %d", int(r.currentGRPCStatus.Code()))
	if cond != "" {
		r.currentGRPCTestCond = append(r.currentGRPCTestCond, cond)
	}
	if len(r.currentGRPCTestCond) == 0 {
		return
	}
	step := r.latestStep()
	step = append(step, yaml.MapItem{Key: "test", Value: fmt.Sprintf("%s\n", strings.Join(r.currentGRPCTestCond, "\n&& "))})
	r.replaceLatestStep(step)
	r.currentGRPCTestCond = nil
	r.currentGRPCResponceIndex = 0
}

func (c *cRunbook) CaptureCDPStart(name string) {
	// FIXME: not implemented
}
func (c *cRunbook) CaptureCDPAction(a runn.CDPAction) {
	// FIXME: not implemented
}
func (c *cRunbook) CaptureCDPResponse(a runn.CDPAction, res map[string]any) {
	// FIXME: not implemented
}
func (c *cRunbook) CaptureCDPEnd(name string) {
	// FIXME: not implemented
}

func (c *cRunbook) CaptureSSHCommand(command string) {
	// FIXME: not implemented
}

func (c *cRunbook) CaptureSSHStdout(stdout string) {
	// FIXME: not implemented
}

func (c *cRunbook) CaptureSSHStderr(stderr string) {
	// FIXME: not implemented
}

func (c *cRunbook) CaptureDBStatement(name string, stmt string) {
	const dummyDsn = "[THIS IS DB RUNNER]"
	if v, ok := c.runners[name]; ok {
		c.setRunner(name, v)
	} else {
		c.setRunner(name, dummyDsn)
	}
	r := c.currentRunbook()
	if r == nil {
		return
	}
	step := yaml.MapSlice{
		{Key: name, Value: yaml.MapSlice{
			{Key: "query", Value: fmt.Sprintf("%s\n", stmt)},
		}},
	}
	r.Steps = append(r.Steps, step)
}

func (c *cRunbook) CaptureDBResponse(name string, res *runn.DBResponse) {
	const threshold = 3

	r := c.currentRunbook()
	if r == nil {
		return
	}
	var cond []string
	if len(res.Columns) > 0 {
		cond = append(cond, fmt.Sprintf("len(current.rows) == %d", len(res.Rows)))
	}
	if len(res.Columns) > 0 && len(res.Rows) <= threshold {
		for i, r := range res.Rows {
			b, err := json.Marshal(r)
			if err != nil {
				c.errs = errors.Join(c.errs, fmt.Errorf("failed to yaml.Marshal: %w", err))
				return
			}
			cond = append(cond, fmt.Sprintf("compare(current.rows[%d], %s)", i, string(b)))
		}
	}
	step := r.latestStep()
	if len(cond) > 0 {
		step = append(step, yaml.MapItem{Key: "test", Value: fmt.Sprintf("%s\n", strings.Join(cond, "\n&& "))})
	}
	r.replaceLatestStep(step)
}

func (c *cRunbook) CaptureExecCommand(command, shell string, background bool) {
	r := c.currentRunbook()
	if r == nil {
		return
	}
	cmd := yaml.MapSlice{
		{Key: "command", Value: command},
		{Key: "shell", Value: shell},
	}
	if background {
		cmd = append(cmd, yaml.MapItem{Key: "background", Value: true})
	}
	step := yaml.MapSlice{
		{Key: "exec", Value: cmd},
	}
	r.Steps = append(r.Steps, step)
}

func (c *cRunbook) CaptureExecStdin(stdin string) {
	if stdin == "" {
		return
	}
	r := c.currentRunbook()
	if r == nil {
		return
	}
	step := r.latestStep()
	exec, ok := step[0].Value.(yaml.MapSlice)
	if !ok {
		c.errs = errors.Join(c.errs, fmt.Errorf("failed to get step[0].Value: %s", step[0].Value))
		return
	}
	exec = append(exec, yaml.MapItem{Key: "stdin", Value: stdin})
	step[0].Value = exec
	r.replaceLatestStep(step)
}

func (c *cRunbook) CaptureExecStdout(stdout string) {
	r := c.currentRunbook()
	if r == nil {
		return
	}
	r.currentExecTestCond = append(r.currentExecTestCond, fmt.Sprintf("current.stdout == %#v", stdout))
}

func (c *cRunbook) CaptureExecStderr(stderr string) {
	r := c.currentRunbook()
	if r == nil {
		return
	}
	r.currentExecTestCond = append(r.currentExecTestCond, fmt.Sprintf("current.stderr == %#v", stderr))
	step := r.latestStep()
	step = append(step, yaml.MapItem{Key: "test", Value: fmt.Sprintf("%s\n", strings.Join(r.currentExecTestCond, "\n&& "))})
	r.replaceLatestStep(step)
	r.currentExecTestCond = nil
}

func (c *cRunbook) SetCurrentTrails(trs runn.Trails) {
	c.currentTrails = trs
}

func (c *cRunbook) Errs() error {
	return c.errs
}

func (c *cRunbook) setRunner(name string, value any) {
	r := c.currentRunbook()
	if r == nil {
		return
	}
	exist := false
	for _, rnr := range r.Runners {
		if rnr.Key.(string) == name {
			exist = true
		}
	}
	if !exist {
		r.Runners = append(r.Runners, yaml.MapItem{Key: name, Value: value})
	}
}

func (c *cRunbook) currentRunbook() *runbook {
	v, ok := c.runbooks.Load(c.currentTrails[0])
	if !ok {
		c.errs = errors.Join(c.errs, fmt.Errorf("failed to c.runbooks.Load: %s", c.currentTrails[0]))
		return nil
	}
	r, ok := v.(*runbook)
	if !ok {
		c.errs = errors.Join(c.errs, fmt.Errorf("failed to cast: %#v", v))
		return nil
	}
	return r
}

func (c *cRunbook) captureGRPCResponseMetadata(key string, m map[string][]string) {
	if len(m) == 0 {
		return
	}
	r := c.currentRunbook()
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		for i, v := range m[k] {
			cond := fmt.Sprintf("current.res.%s[%q][%d] == %#v", key, k, i, v)
			r.currentGRPCTestCond = append(r.currentGRPCTestCond, cond)
		}
	}
}

func (c *cRunbook) appendOp(hb yaml.MapSlice, m any) yaml.MapSlice {
	switch {
	case len(hb) == 0 || (len(hb) == 1 && hb[0].Key == "headers"):
		hb = append(hb, yaml.MapItem{Key: "messages", Value: []any{m}})
	case hb[0].Key == "messages":
		ms, ok := hb[0].Value.([]any)
		if !ok {
			c.errs = errors.Join(c.errs, fmt.Errorf("failed to get hb[0].Value: %s", hb[0].Value))
			return hb
		}
		ms = append(ms, m)
		hb[0].Value = ms
	case hb[1].Key == "messages":
		ms, ok := hb[1].Value.([]any)
		if !ok {
			c.errs = errors.Join(c.errs, fmt.Errorf("failed to get hb[1].Value: %s", hb[1].Value))
			return hb
		}
		ms = append(ms, m)
		hb[1].Value = ms
	}
	return hb
}

func (c *cRunbook) writeRunbook(trs runn.Trails, bookPath string) {
	v, ok := c.runbooks.Load(trs[0])
	if !ok {
		c.errs = errors.Join(c.errs, fmt.Errorf("failed to c.runbooks.Load: %s", trs[0]))
		return
	}
	r, ok := v.(*runbook)
	if !ok {
		c.errs = errors.Join(c.errs, fmt.Errorf("failed to cast: %#v", v))
		return
	}
	if c.desc != "" {
		r.Desc = c.desc
	} else {
		r.Desc = fmt.Sprintf("Captured of %s run", filepath.Base(bookPath))
	}
	r.Labels = c.labels
	b, err := yaml.Marshal(r)
	if err != nil {
		c.errs = errors.Join(c.errs, fmt.Errorf("failed to yaml.Marshal: %w", err))
		return
	}
	p := filepath.Join(c.dir, capturedFilename(bookPath))
	if err := os.WriteFile(p, b, os.ModePerm); err != nil { //nolint:gosec
		c.errs = errors.Join(c.errs, err)
		return
	}
}

func (r *runbook) latestStep() yaml.MapSlice {
	return r.Steps[len(r.Steps)-1]
}

func (r *runbook) replaceLatestStep(rep yaml.MapSlice) {
	r.Steps[len(r.Steps)-1] = rep
}

func headersAndMessages(step yaml.MapSlice) yaml.MapSlice {
	return step[0].Value.(yaml.MapSlice)[0].Value.(yaml.MapSlice)
}

func replaceHeadersAndMessages(step, hb yaml.MapSlice) yaml.MapSlice {
	step[0].Value.(yaml.MapSlice)[0].Value = hb
	return step
}

// copy from net/http/httputil.
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == nil || b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return io.NopCloser(&buf), io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func capturedFilename(bookPath string) string {
	return strings.ReplaceAll(strings.ReplaceAll(bookPath, string(filepath.Separator), "-"), "..", "")
}
