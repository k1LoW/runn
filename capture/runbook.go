package capture

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	Desc    string                   `yaml:"desc"`
	Runners map[string]interface{}   `yaml:"runners,omitempty"`
	Steps   []map[string]interface{} `yaml:"steps"`
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

func (c *cRunbook) CaptureHTTPRequest(req *http.Request) {}

func (c *cRunbook) CaptureHTTPResponse(res *http.Response) {}

func (c *cRunbook) CaptureGRPCStart(service, method string) {}

func (c *cRunbook) CaptureGRPCRequestHeaders(h map[string][]string) {}

func (c *cRunbook) CaptureGRPCRequestMessage(m map[string]interface{}) {}

func (c *cRunbook) CaptureGRPCResponseStatus(status int) {}

func (c *cRunbook) CaptureGRPCResponseHeaders(h map[string][]string) {}

func (c *cRunbook) CaptureGRPCResponseMessage(m map[string]interface{}) {}

func (c *cRunbook) CaptureGRPCResponseTrailers(t map[string][]string) {}

func (c *cRunbook) CaptureGRPCEnd(service, method string) {}

func (c *cRunbook) CaptureDBStatement(stmt string) {}

func (c *cRunbook) CaptureDBResponse(res *runn.DBResponse) {}

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
