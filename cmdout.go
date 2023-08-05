package runn

import (
	"errors"
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
			_, _ = fmt.Fprintf(d.out, "%s", red("F"))
		case result.Skipped:
			_, _ = fmt.Fprintf(d.out, "%s", yellow("S"))
		default:
			_, _ = fmt.Fprintf(d.out, "%s", green("."))
		}
		return
	}
	switch {
	case result.Err != nil:
		_, _ = fmt.Fprintf(d.out, "=== %s (%s) ... %v\n", result.Desc, ShortenPath(result.Path), red("fail"))
	case result.Skipped:
		_, _ = fmt.Fprintf(d.out, "=== %s (%s) ... %s\n", result.Desc, ShortenPath(result.Path), yellow("skip"))
	default:
		_, _ = fmt.Fprintf(d.out, "=== %s (%s) ... %s\n", result.Desc, ShortenPath(result.Path), green("ok"))
	}
	for _, sr := range result.StepResults {
		desc := ""
		if sr.Desc != "" {
			desc = fmt.Sprintf("%s ", sr.Desc)
		}
		switch {
		case sr.Err != nil:
			uerr := errors.Unwrap(sr.Err)
			if uerr == nil {
				uerr = sr.Err
			}
			_, _ = fmt.Fprintf(d.out, "    --- %s(%s) ... %s\n%s\n", desc, sr.Key, red("fail"), red(SprintMultilinef("        %s\n", "Failure/Error: %s", strings.TrimRight(uerr.Error(), "\n"))))
		case sr.Skipped:
			_, _ = fmt.Fprintf(d.out, "    --- %s(%s) ... %s\n", desc, sr.Key, yellow("skip"))
		default:
			_, _ = fmt.Fprintf(d.out, "    --- %s(%s) ... %s\n", desc, sr.Key, green("ok"))
		}
	}
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
