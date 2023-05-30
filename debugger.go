package runn

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"sort"
	"strings"

	"github.com/goccy/go-json"
	"github.com/olekukonko/tablewriter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ Capturer = (*debugger)(nil)

type debugger struct {
	out        io.Writer
	currentIDs IDs
	errs       error
}

func NewDebugger(out io.Writer) *debugger {
	return &debugger{
		out: out,
	}
}

func (d *debugger) CaptureStart(ids IDs, bookPath, desc string) {}
func (d *debugger) CaptureResult(ids IDs, result *RunResult)    {}
func (d *debugger) CaptureEnd(ids IDs, bookPath, desc string)   {}

func (d *debugger) CaptureHTTPRequest(name string, req *http.Request) {
	b, _ := httputil.DumpRequest(req, true)
	_, _ = fmt.Fprintf(d.out, "-----START HTTP REQUEST-----\n%s\n-----END HTTP REQUEST-----\n", string(b))
}

func (d *debugger) CaptureHTTPResponse(name string, res *http.Response) {
	b, _ := httputil.DumpResponse(res, true)
	_, _ = fmt.Fprintf(d.out, "-----START HTTP RESPONSE-----\n%s\n-----END HTTP RESPONSE-----\n", string(b))
}

func (d *debugger) CaptureGRPCStart(name string, typ GRPCType, service, method string) {
	_, _ = fmt.Fprintf(d.out, ">>>>>START gRPC (%s/%s)>>>>>\n", service, method)
}

func (d *debugger) CaptureGRPCRequestHeaders(h map[string][]string) {
	_, _ = fmt.Fprintf(d.out, "-----START gRPC REQUEST HEADERS-----\n%s\n-----END gRPC REQUEST HEADERS-----\n", dumpGRPCMetadata(h))
}

func (d *debugger) CaptureGRPCRequestMessage(m map[string]interface{}) {
	_, _ = fmt.Fprintf(d.out, "-----START gRPC REQUEST MESSAGE-----\n%s\n-----END gRPC REQUEST MESSAGE-----\n", dumpGRPCMessage(m))
}

func (d *debugger) CaptureGRPCResponseStatus(s *status.Status) {
	c := s.Code()
	m := fmt.Sprintf("%s (%d)", c.String(), int(c))
	if c != codes.OK {
		m = fmt.Sprintf("%s (%d): %s", c.String(), int(c), s.Message())
	}
	_, _ = fmt.Fprintf(d.out, "-----START gRPC RESPONSE STATUS-----\n%s\n-----END gRPC RESPONSE STATUS-----\n", m)
}

func (d *debugger) CaptureGRPCResponseHeaders(h map[string][]string) {
	_, _ = fmt.Fprintf(d.out, "-----START gRPC RESPONSE HEADERS-----\n%s\n-----END gRPC RESPONSE HEADERS-----\n", dumpGRPCMetadata(h))
}

func (d *debugger) CaptureGRPCResponseMessage(m map[string]interface{}) {
	_, _ = fmt.Fprintf(d.out, "-----START gRPC RESPONSE MESSAGE-----\n%s\n-----END gRPC RESPONSE MESSAGE-----\n", dumpGRPCMessage(m))
}

func (d *debugger) CaptureGRPCResponseTrailers(t map[string][]string) {
	_, _ = fmt.Fprintf(d.out, "-----START gRPC RESPONSE TRAILERS-----\n%s\n-----END gRPC RESPONSE TRAILERS-----\n", dumpGRPCMetadata(t))
}

func (d *debugger) CaptureGRPCClientClose() {}

func (d *debugger) CaptureGRPCEnd(name string, typ GRPCType, service, method string) {
	_, _ = fmt.Fprintf(d.out, "<<<<<END gRPC (%s/%s)<<<<<\n", service, method)
}

func (d *debugger) CaptureCDPStart(name string) {
	_, _ = fmt.Fprint(d.out, ">>>>>START CDP>>>>>\n")
}
func (d *debugger) CaptureCDPAction(a CDPAction) {
	_, _ = fmt.Fprintf(d.out, "-----START CDP ACTION-----\nname: %s\nargs:\n%s\n-----END CDP ACTION-----\n", a.Fn, dumpCDPValues(a.Args))
}
func (d *debugger) CaptureCDPResponse(a CDPAction, res map[string]interface{}) {
	_, _ = fmt.Fprintf(d.out, "-----START CDP RESPONSE-----\nname: %s\nresponse:\n%s\n-----END CDP RESPONSE-----\n", a.Fn, dumpCDPValues(res))
}
func (d *debugger) CaptureCDPEnd(name string) {
	_, _ = fmt.Fprint(d.out, "<<<<<END CDP<<<<<\n")
}

func (d *debugger) CaptureSSHCommand(command string) {
	_, _ = fmt.Fprintf(d.out, "-----START COMMAND-----\n%s\n-----END COMMAND-----\n", command)
}

func (d *debugger) CaptureSSHStdout(stdout string) {
	_, _ = fmt.Fprintf(d.out, "-----START STDOUT-----\n%s\n-----END STDOUT-----\n", stdout)
}

func (d *debugger) CaptureSSHStderr(stderr string) {
	_, _ = fmt.Fprintf(d.out, "-----START STDERR-----\n%s\n-----END STDERR-----\n", stderr)
}

func (d *debugger) CaptureDBStatement(name string, stmt string) {
	_, _ = fmt.Fprintf(d.out, "-----START QUERY-----\n%s\n-----END QUERY-----\n", stmt)
}

func (d *debugger) CaptureDBResponse(name string, res *DBResponse) {
	_, _ = fmt.Fprint(d.out, "-----START QUERY RESULT-----\n")
	defer fmt.Fprint(d.out, "-----END QUERY RESULT-----\n")
	if len(res.Rows) == 0 {
		_, _ = fmt.Fprintf(d.out, "rows affected: %d\n", res.RowsAffected)
		if res.LastInsertID > 0 {
			_, _ = fmt.Fprintf(d.out, "last insert id: %d\n", res.LastInsertID)
		}
		return
	}
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
}

func (d *debugger) CaptureExecCommand(command string) {
	_, _ = fmt.Fprintf(d.out, "-----START COMMAND-----\n%s\n-----END COMMAND-----\n", command)
}

func (d *debugger) CaptureExecStdin(stdin string) {
	_, _ = fmt.Fprintf(d.out, "-----START STDIN-----\n%s\n-----END STDIN-----\n", stdin)
}

func (d *debugger) CaptureExecStdout(stdout string) {
	_, _ = fmt.Fprintf(d.out, "-----START STDOUT-----\n%s\n-----END STDOUT-----\n", stdout)
}

func (d *debugger) CaptureExecStderr(stderr string) {
	_, _ = fmt.Fprintf(d.out, "-----START STDERR-----\n%s\n-----END STDERR-----\n", stderr)
}

func (d *debugger) SetCurrentIDs(ids IDs) {
	d.currentIDs = ids
}

func (d *debugger) Errs() error {
	return d.errs
}

func dumpMapInterface(m map[string]interface{}) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	var d []string
	for _, k := range keys {
		switch v := m[k].(type) {
		case string:
			d = append(d, fmt.Sprintf(`%s: %#v`, k, v))
		default:
			b, _ := json.Marshal(v)
			d = append(d, fmt.Sprintf(`%s: %s`, k, string(b)))
		}
	}
	return strings.Join(d, "\n")
}

var (
	dumpCDPValues   = dumpMapInterface
	dumpGRPCMessage = dumpMapInterface
)

func dumpGRPCMetadata(m map[string][]string) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	var d []string
	for _, k := range keys {
		b, _ := json.Marshal(m[k])
		d = append(d, fmt.Sprintf(`%s: %s`, k, string(b)))
	}
	return strings.Join(d, "\n")
}
