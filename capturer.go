package runn

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc/metadata"
)

type Capturer interface {
	CaptureHTTPRequest(req *http.Request)
	CaptureHTTPResponse(res *http.Response)

	CaptureGRPCStart()
	CaptureGRPCRequestHeaders(h map[string][]string)
	CaptureGRPCRequestMessage(m map[string]interface{})
	CaptureGRPCResponseStatus(status int)
	CaptureGRPCResponseHeaders(h map[string][]string)
	CaptureGRPCResponseMessage(m map[string]interface{})
	CaptureGRPCResponseTrailers(t map[string][]string)
	CaptureGRPCEnd()

	CaptureDBStatement(stmt string)
	CaptureDBResponse(res *DBResponse)
	CaptureExecCommand(command string)
	CaptureExecStdin(stdin string)
	CaptureExecStdout(stdin string)
	CaptureExecStderr(stderr string)
	SetCurrentIDs(ids []string)
	Errs() error
}

type capturers []Capturer

func (cs capturers) captureHTTPRequest(req *http.Request) {
	for _, c := range cs {
		c.CaptureHTTPRequest(req)
	}
}

func (cs capturers) captureHTTPResponse(res *http.Response) {
	for _, c := range cs {
		c.CaptureHTTPResponse(res)
	}
}

func (cs capturers) captureGRPCStart() {
	for _, c := range cs {
		c.CaptureGRPCStart()
	}
}
func (cs capturers) captureGRPCRequestHeaders(h metadata.MD) {
	for _, c := range cs {
		c.CaptureGRPCRequestHeaders(h)
	}
}

func (cs capturers) captureGRPCRequestMessage(m map[string]interface{}) {
	for _, c := range cs {
		c.CaptureGRPCRequestMessage(m)
	}
}

func (cs capturers) captureGRPCResponseStatus(status int) {
	for _, c := range cs {
		c.CaptureGRPCResponseStatus(status)
	}
}

func (cs capturers) captureGRPCResponseHeaders(h metadata.MD) {
	for _, c := range cs {
		c.CaptureGRPCResponseHeaders(h)
	}
}

func (cs capturers) captureGRPCResponseMessage(m map[string]interface{}) {
	for _, c := range cs {
		c.CaptureGRPCResponseMessage(m)
	}
}

func (cs capturers) captureGRPCResponseTrailers(t metadata.MD) {
	for _, c := range cs {
		c.CaptureGRPCResponseTrailers(t)
	}
}

func (cs capturers) captureGRPCEnd() {
	for _, c := range cs {
		c.CaptureGRPCEnd()
	}
}

func (cs capturers) captureDBStatement(stmt string) {
	for _, c := range cs {
		c.CaptureDBStatement(stmt)
	}
}

func (cs capturers) captureDBResponse(res *DBResponse) {
	for _, c := range cs {
		c.CaptureDBResponse(res)
	}
}

func (cs capturers) captureExecCommand(command string) {
	for _, c := range cs {
		c.CaptureExecCommand(command)
	}
}

func (cs capturers) captureExecStdin(stdin string) {
	for _, c := range cs {
		c.CaptureExecStdin(stdin)
	}
}

func (cs capturers) captureExecStdout(stdout string) {
	for _, c := range cs {
		c.CaptureExecStdout(stdout)
	}
}

func (cs capturers) captureExecStderr(stderr string) {
	for _, c := range cs {
		c.CaptureExecStderr(stderr)
	}
}

func (cs capturers) setCurrentIDs(ids []string) {
	for _, c := range cs {
		c.SetCurrentIDs(ids)
	}
}

var _ Capturer = (*debugger)(nil)

type debugger struct {
	out        io.Writer
	currentIDs []string
	errs       error
}

func NewDebugger(out io.Writer) *debugger {
	return &debugger{
		out: out,
	}
}

func (d *debugger) CaptureHTTPRequest(req *http.Request) {
	b, _ := httputil.DumpRequest(req, true)
	_, _ = fmt.Fprintf(d.out, "-----START HTTP REQUEST-----\n%s\n-----END HTTP REQUEST-----\n", string(b))
}

func (d *debugger) CaptureHTTPResponse(res *http.Response) {
	b, _ := httputil.DumpResponse(res, true)
	_, _ = fmt.Fprintf(d.out, "-----START HTTP RESPONSE-----\n%s\n-----END HTTP RESPONSE-----\n", string(b))
}

func (d *debugger) CaptureGRPCStart() {
	_, _ = fmt.Fprint(d.out, "-----START gRPC-----\n")
}

func (d *debugger) CaptureGRPCRequestHeaders(h map[string][]string)     {}
func (d *debugger) CaptureGRPCRequestMessage(m map[string]interface{})  {}
func (d *debugger) CaptureGRPCResponseStatus(status int)                {}
func (d *debugger) CaptureGRPCResponseHeaders(h map[string][]string)    {}
func (d *debugger) CaptureGRPCResponseMessage(m map[string]interface{}) {}
func (d *debugger) CaptureGRPCResponseTrailers(t map[string][]string)   {}
func (d *debugger) CaptureGRPCEnd() {
	_, _ = fmt.Fprint(d.out, "-----END gRPC-----\n")
}

func (d *debugger) CaptureDBStatement(stmt string) {
	_, _ = fmt.Fprintf(d.out, "-----START QUERY-----\n%s\n-----END QUERY-----\n", stmt)
}

func (d *debugger) CaptureDBResponse(res *DBResponse) {
	if len(res.Rows) == 0 {
		return
	}
	_, _ = fmt.Fprint(d.out, "-----START ROWS-----\n")
	table := tablewriter.NewWriter(d.out)
	table.SetHeader(res.Columns)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	for _, r := range res.Rows {
		row := make([]string, 0, len(res.Columns))
		for _, c := range res.Columns {
			row = append(row, fmt.Sprintf("%v", r[c]))
		}
		table.Append(row)
	}
	table.Render()
	c := len(res.Rows)
	if c == 1 {
		_, _ = fmt.Fprintf(d.out, "(%d row)\n", len(res.Rows))
	} else {
		_, _ = fmt.Fprintf(d.out, "(%d rows)\n", len(res.Rows))
	}
	_, _ = fmt.Fprint(d.out, "-----END ROWS-----\n")
}

func (d *debugger) CaptureExecCommand(command string) {
	_, _ = fmt.Fprintf(d.out, "-----START COMMAND-----\n%s\n-----END COMMAND-----\n", command)
}

func (d *debugger) CaptureExecStdin(stdin string) {
	_, _ = fmt.Fprintf(d.out, "-----START STDIN-----\n%s\n-----END STDIN-----\n", stdin)
}

func (d *debugger) CaptureExecStdout(stdout string) {
	_, _ = fmt.Fprintf(d.out, "-----START STDIN-----\n%s\n-----END STDIN-----\n", stdout)
}

func (d *debugger) CaptureExecStderr(stderr string) {
	_, _ = fmt.Fprintf(d.out, "-----START STDERR-----\n%s\n-----END STDERR-----\n", stderr)
}

func (d *debugger) SetCurrentIDs(ids []string) {
	d.currentIDs = ids
}

func (d *debugger) Errs() error {
	return d.errs
}
