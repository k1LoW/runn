package capture

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/k1LoW/runn"
	"gopkg.in/yaml.v2"
)

var _ runn.Capturer = (*cRunbook)(nil)

type cRunbook struct {
	dir        string
	currentIDs []string
	errs       error
	runbooks   sync.Map
}

type runbook struct {
	Desc    string          `yaml:"desc"`
	Runners yaml.MapSlice   `yaml:"runners,omitempty"`
	Steps   []yaml.MapSlice `yaml:"steps"`
}

func Runbook(dir string) *cRunbook {
	return &cRunbook{
		dir:      dir,
		runbooks: sync.Map{},
	}
}

func (c *cRunbook) CaptureStart(ids []string, bookPath string) {
	c.runbooks.Store(ids[0], &runbook{})
}

func (c *cRunbook) CaptureEnd(ids []string, bookPath string) {
	v, ok := c.runbooks.Load(ids[0])
	if !ok {
		return
	}
	r, ok := v.(*runbook)
	if !ok {
		return
	}
	r.Desc = fmt.Sprintf("Captured of %s run", filepath.Base(bookPath))
	b, err := yaml.Marshal(r)
	if err != nil {
		return
	}
	p := filepath.Join(c.dir, capturedFilename(bookPath))
	_ = os.WriteFile(p, b, os.ModePerm)
}

func (c *cRunbook) CaptureHTTPRequest(name string, req *http.Request) {
	c.setRunner(name, "[THIS IS HTTP RUNNER]")
	r := c.currentRunbook()
	if r == nil {
		return
	}
	endpoint := req.URL.Path
	if req.URL.RawQuery != "" {
		endpoint = fmt.Sprintf("%s?%s", endpoint, req.URL.RawQuery)
	}

	hb := yaml.MapSlice{}
	// headers
	contentType := req.Header.Get("Content-Type")
	h := map[string]string{}
	for k, v := range req.Header {
		if k == "Content-Type" || k == "Host" {
			continue
		}
		h[k] = v[0]
	}
	if len(h) > 0 {
		hb = append(hb, yaml.MapItem{
			Key:   "headers",
			Value: h,
		})
	}

	// body
	var bd yaml.MapSlice
	var (
		save io.ReadCloser
		err  error
	)
	save, req.Body, err = drainBody(req.Body)
	if err != nil {
		return
	}
	switch {
	case save == http.NoBody || save == nil:
		bd = yaml.MapSlice{
			{Key: contentType, Value: nil},
		}
	case strings.Contains(contentType, "json"):
		var v interface{}
		if err := json.NewDecoder(save).Decode(&v); err != nil {
			return
		}
		bd = yaml.MapSlice{
			{Key: contentType, Value: v},
		}
	case contentType == runn.MediaTypeApplicationFormUrlencoded:
		b, err := io.ReadAll(save)
		if err != nil {
			return
		}
		vs, err := url.ParseQuery(string(b))
		if err != nil {
			return
		}
		f := map[string]interface{}{}
		for k, v := range vs {
			if len(v) == 1 {
				f[k] = v[0]
				continue
			}
			f[k] = v
		}
		bd = yaml.MapSlice{
			{Key: contentType, Value: f},
		}
	default:
		// case contentType == runn.MediaTypeTextPlain:
		b, err := io.ReadAll(save)
		if err != nil {
			return
		}
		bd = yaml.MapSlice{
			{Key: contentType, Value: string(b)},
		}
	}
	hb = append(hb, yaml.MapItem{
		Key:   "body",
		Value: bd,
	})

	m := yaml.MapItem{Key: strings.ToLower(req.Method), Value: nil}
	if len(hb) > 0 {
		m = yaml.MapItem{Key: strings.ToLower(req.Method), Value: hb}
	}

	step := yaml.MapSlice{
		{Key: name, Value: yaml.MapSlice{
			{Key: endpoint, Value: yaml.MapSlice{
				m,
			}},
		}},
	}
	r.Steps = append(r.Steps, step)
}

func (c *cRunbook) CaptureHTTPResponse(name string, res *http.Response) {
	r := c.currentRunbook()
	step := r.latestStep()
	// status
	cond := fmt.Sprintf("current.res.status == %d\n", res.StatusCode)

	// headers
	keys := []string{}
	for k := range res.Header {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if k == "Date" {
			continue
		}
		for i, v := range res.Header[k] {
			cond += fmt.Sprintf("&& current.res.headers['%s'][%d] == '%s'\n", k, i, v)
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
		return
	}
	if strings.Contains(contentType, "json") {

	} else {
		b, err := io.ReadAll(save)
		if err != nil {
			return
		}
		cond += fmt.Sprintf("&& current.res.rawBody == %#v\n", string(b))
	}

	r.replaceLatestStep(append(step, yaml.MapItem{Key: "test", Value: cond}))
}

func (c *cRunbook) CaptureGRPCStart(name string, service, method string) {
	c.setRunner(name, "[THIS IS gRPC RUNNER]")
}

func (c *cRunbook) CaptureGRPCRequestHeaders(h map[string][]string) {}

func (c *cRunbook) CaptureGRPCRequestMessage(m map[string]interface{}) {}

func (c *cRunbook) CaptureGRPCResponseStatus(status int) {}

func (c *cRunbook) CaptureGRPCResponseHeaders(h map[string][]string) {}

func (c *cRunbook) CaptureGRPCResponseMessage(m map[string]interface{}) {}

func (c *cRunbook) CaptureGRPCResponseTrailers(t map[string][]string) {}

func (c *cRunbook) CaptureGRPCEnd(name string, service, method string) {}

func (c *cRunbook) CaptureDBStatement(name string, stmt string) {
	c.setRunner(name, "[THIS IS DB RUNNER]")
}

func (c *cRunbook) CaptureDBResponse(name string, res *runn.DBResponse) {}

func (c *cRunbook) CaptureExecCommand(command string) {}

func (c *cRunbook) CaptureExecStdin(stdin string) {}

func (c *cRunbook) CaptureExecStdout(stdout string) {}

func (c *cRunbook) CaptureExecStderr(stderr string) {}

func (c *cRunbook) SetCurrentIDs(ids []string) {
	c.currentIDs = ids
}

func (c *cRunbook) Errs() error {
	return c.errs
}

func (c *cRunbook) setRunner(name, value string) {
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
	v, ok := c.runbooks.Load(c.currentIDs[0])
	if !ok {
		return nil
	}
	r, ok := v.(*runbook)
	if !ok {
		return nil
	}
	return r
}

func (r *runbook) latestStep() yaml.MapSlice {
	return r.Steps[len(r.Steps)-1]
}

func (r *runbook) replaceLatestStep(rep yaml.MapSlice) {
	r.Steps[len(r.Steps)-1] = rep
}

// copy from net/http/httputil
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
