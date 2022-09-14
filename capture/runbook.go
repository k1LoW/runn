package capture

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/runn"
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
	p := filepath.Join(c.dir, strings.ReplaceAll(strings.ReplaceAll(bookPath, string(filepath.Separator), "-"), "..", ""))
	_ = os.WriteFile(p, b, os.ModePerm)
}

func (c *cRunbook) CaptureHTTPRequest(name string, req *http.Request) {
	c.setRunner(name, "[THIS IS HTTP RUNNER]")
	r := c.loadCurrentRunbook()
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
	if contentType == "" {
		return
	}
	h := yaml.MapSlice{}
	for k, v := range req.Header {
		if k == "Content-Type" || k == "Host" {
			continue
		}
		h = append(h, yaml.MapItem{
			Key:   k,
			Value: v[0],
		})
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

func (c *cRunbook) CaptureHTTPResponse(name string, res *http.Response) {}

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
	r := c.loadCurrentRunbook()
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

func (c *cRunbook) loadCurrentRunbook() *runbook {
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
