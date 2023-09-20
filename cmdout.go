package runn

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/grpc/status"
)

var _ Capturer = (*cmdOut)(nil)

type cmdOut struct {
	out     io.Writer
	verbose bool
	errs    error
}

func NewCmdOut(out io.Writer, verbose bool) *cmdOut {
	return &cmdOut{
		out:     out,
		verbose: verbose,
	}
}

func (d *cmdOut) CaptureStart(trs Trails, bookPath, desc string) {}
func (d *cmdOut) CaptureResult(trs Trails, result *RunResult) {
	if !d.verbose {
		switch {
		case result.Err != nil:
			_, _ = fmt.Fprint(d.out, red("F"))
		case result.Skipped:
			_, _ = fmt.Fprint(d.out, yellow("S"))
		default:
			_, _ = fmt.Fprint(d.out, green("."))
		}
		return
	}
	// verbose
	d.verboseOutResult(result, 0)
}
func (d *cmdOut) CaptureEnd(trs Trails, bookPath, desc string) {}

func (d *cmdOut) CaptureHTTPRequest(name string, req *http.Request)                  {}
func (d *cmdOut) CaptureHTTPResponse(name string, res *http.Response)                {}
func (d *cmdOut) CaptureGRPCStart(name string, typ GRPCType, service, method string) {}
func (d *cmdOut) CaptureGRPCRequestHeaders(h map[string][]string)                    {}
func (d *cmdOut) CaptureGRPCRequestMessage(m map[string]any)                         {}
func (d *cmdOut) CaptureGRPCResponseStatus(s *status.Status)                         {}
func (d *cmdOut) CaptureGRPCResponseHeaders(h map[string][]string)                   {}
func (d *cmdOut) CaptureGRPCResponseMessage(m map[string]any)                        {}
func (d *cmdOut) CaptureGRPCResponseTrailers(t map[string][]string)                  {}
func (d *cmdOut) CaptureGRPCClientClose()                                            {}
func (d *cmdOut) CaptureGRPCEnd(name string, typ GRPCType, service, method string)   {}
func (d *cmdOut) CaptureCDPStart(name string)                                        {}
func (d *cmdOut) CaptureCDPAction(a CDPAction)                                       {}
func (d *cmdOut) CaptureCDPResponse(a CDPAction, res map[string]any)                 {}
func (d *cmdOut) CaptureCDPEnd(name string)                                          {}
func (d *cmdOut) CaptureSSHCommand(command string)                                   {}
func (d *cmdOut) CaptureSSHStdout(stdout string)                                     {}
func (d *cmdOut) CaptureSSHStderr(stderr string)                                     {}
func (d *cmdOut) CaptureDBStatement(name string, stmt string)                        {}
func (d *cmdOut) CaptureDBResponse(name string, res *DBResponse)                     {}
func (d *cmdOut) CaptureExecCommand(command string)                                  {}
func (d *cmdOut) CaptureExecStdin(stdin string)                                      {}
func (d *cmdOut) CaptureExecStdout(stdout string)                                    {}
func (d *cmdOut) CaptureExecStderr(stderr string)                                    {}
func (d *cmdOut) SetCurrentTrails(trs Trails)                                        {}
func (d *cmdOut) Errs() error {
	return d.errs
}

func (d *cmdOut) verboseOutResult(r *RunResult, idx int) {
	// verbose
	indent := strings.Repeat("        ", idx)
	switch {
	case r.Err != nil:
		_, _ = fmt.Fprintf(d.out, "%s=== %s (%s) ... %v\n", indent, r.Desc, r.Path, red("fail"))
	case r.Skipped:
		_, _ = fmt.Fprintf(d.out, "%s=== %s (%s) ... %s\n", indent, r.Desc, r.Path, yellow("skip"))
	default:
		_, _ = fmt.Fprintf(d.out, "%s=== %s (%s) ... %s\n", indent, r.Desc, r.Path, green("ok"))
	}
	for i, sr := range r.StepResults {
		desc := ""
		if sr.Desc != "" {
			desc = fmt.Sprintf("%s ", sr.Desc)
		}
		switch {
		case sr.Err != nil:
			if sr.IncludedRunResult != nil {
				_, _ = fmt.Fprintf(d.out, "%s    --- %s(%s) ... %s\n", indent, desc, sr.Key, red("fail"))
				d.verboseOutResult(sr.IncludedRunResult, idx+1)
				continue
			}
			lineformat := indent + "        %s\n"
			_, _ = fmt.Fprintf(d.out, "%s    --- %s(%s) ... %s\n%s", indent, desc, sr.Key, red("fail"), red(SprintMultilinef(lineformat, "Failure/Error: %s", strings.TrimRight(sr.Err.Error(), "\n"))))
			b, err := readFile(r.Path)
			if err != nil {
				continue
			}
			picked, err := pickStepYAML(string(b), i)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(d.out, "%s        Failure step (%s):\n", indent, r.Path)
			_, _ = fmt.Fprint(d.out, SprintMultilinef(lineformat, "%v", picked))
			_, _ = fmt.Fprintln(d.out, "")
		case sr.Skipped:
			_, _ = fmt.Fprintf(d.out, "%s    --- %s(%s) ... %s\n", indent, desc, sr.Key, yellow("skip"))
			if sr.IncludedRunResult != nil {
				d.verboseOutResult(sr.IncludedRunResult, idx+1)
				continue
			}
		default:
			_, _ = fmt.Fprintf(d.out, "%s    --- %s(%s) ... %s\n", indent, desc, sr.Key, green("ok"))
			if sr.IncludedRunResult != nil {
				d.verboseOutResult(sr.IncludedRunResult, idx+1)
				continue
			}
		}
	}
}
