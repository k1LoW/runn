package runn

import (
	"fmt"
	"io"
	"net/http"

	"github.com/fatih/color"
)

var _ Capturer = (*cmdOut)(nil)

type cmdOut struct {
	out    io.Writer
	errs   error
	green  func(a ...interface{}) string
	yellow func(a ...interface{}) string
	red    func(a ...interface{}) string
}

func NewCmdOut(out io.Writer) *cmdOut {
	return &cmdOut{
		out:    out,
		green:  color.New(color.FgGreen).SprintFunc(),
		yellow: color.New(color.FgYellow).SprintFunc(),
		red:    color.New(color.FgRed).SprintFunc(),
	}
}

func (d *cmdOut) CaptureStart(ids IDs, bookPath, desc string) {}
func (d *cmdOut) CaptureFailed(ids IDs, bookPath, desc string, err error) {
	_, _ = fmt.Fprintf(d.out, "%s ... %v\n", desc, d.red(err))
}
func (d *cmdOut) CaptureSkipped(ids IDs, bookPath, desc string) {
	_, _ = fmt.Fprintf(d.out, "%s ... %s\n", desc, d.yellow("skip"))
}
func (d *cmdOut) CaptureSuccess(ids IDs, bookPath, desc string) {
	_, _ = fmt.Fprintf(d.out, "%s ... %s\n", desc, d.green("ok"))
}
func (d *cmdOut) CaptureEnd(ids IDs, bookPath, desc string) {}

func (d *cmdOut) CaptureHTTPRequest(name string, req *http.Request)                  {}
func (d *cmdOut) CaptureHTTPResponse(name string, res *http.Response)                {}
func (d *cmdOut) CaptureGRPCStart(name string, typ GRPCType, service, method string) {}
func (d *cmdOut) CaptureGRPCRequestHeaders(h map[string][]string)                    {}
func (d *cmdOut) CaptureGRPCRequestMessage(m map[string]interface{})                 {}
func (d *cmdOut) CaptureGRPCResponseStatus(status int)                               {}
func (d *cmdOut) CaptureGRPCResponseHeaders(h map[string][]string)                   {}
func (d *cmdOut) CaptureGRPCResponseMessage(m map[string]interface{})                {}
func (d *cmdOut) CaptureGRPCResponseTrailers(t map[string][]string)                  {}
func (d *cmdOut) CaptureGRPCClientClose()                                            {}
func (d *cmdOut) CaptureGRPCEnd(name string, typ GRPCType, service, method string)   {}
func (d *cmdOut) CaptureDBStatement(name string, stmt string)                        {}
func (d *cmdOut) CaptureDBResponse(name string, res *DBResponse)                     {}
func (d *cmdOut) CaptureExecCommand(command string)                                  {}
func (d *cmdOut) CaptureExecStdin(stdin string)                                      {}
func (d *cmdOut) CaptureExecStdout(stdout string)                                    {}
func (d *cmdOut) CaptureExecStderr(stderr string)                                    {}
func (d *cmdOut) SetCurrentIDs(ids IDs)                                              {}
func (d *cmdOut) Errs() error {
	return d.errs
}
