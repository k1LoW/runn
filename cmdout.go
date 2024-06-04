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

func (d *cmdOut) CaptureStart(trs Trails, bookPath, desc string) {
	if !d.verbose {
		return
	}
	_, _ = fmt.Fprintf(d.out, "=== %s (%s)\n", desc, bookPath)
}
func (d *cmdOut) CaptureResult(trs Trails, result *RunResult) {
	if d.verbose {
		return
	}
	switch {
	case result.Err != nil:
		_, _ = fmt.Fprint(d.out, red("F"))
	case result.Skipped:
		_, _ = fmt.Fprint(d.out, yellow("S"))
	default:
		_, _ = fmt.Fprint(d.out, green("."))
	}
}
func (d *cmdOut) CaptureEnd(trs Trails, bookPath, desc string) {}

func (d *cmdOut) CaptureResultByStep(trs Trails, result *RunResult) {
	if !d.verbose {
		return
	}
	if result.included {
		return
	}
	d.verboseOutResult(result, 0)
}

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
func (d *cmdOut) CaptureExecCommand(command, shell string, background bool)          {}
func (d *cmdOut) CaptureExecStdin(stdin string)                                      {}
func (d *cmdOut) CaptureExecStdout(stdout string)                                    {}
func (d *cmdOut) CaptureExecStderr(stderr string)                                    {}
func (d *cmdOut) SetCurrentTrails(trs Trails)                                        {}
func (d *cmdOut) Errs() error {
	return d.errs
}

func (d *cmdOut) verboseOutResult(r *RunResult, nest int) {
	if nest == 0 {
		idx := len(r.StepResults) - 1
		for i, sr := range r.StepResults {
			if sr == nil {
				idx = i - 1
				break
			}
		}
		if idx < 0 {
			return
		}
		sr := r.StepResults[idx]
		d.verboseOutResultForStep(idx, sr, r.Path, nest)
	} else {
		for idx, sr := range r.StepResults {
			if sr == nil {
				continue
			}
			d.verboseOutResultForStep(idx, sr, r.Path, nest)
		}
	}
}

func (d *cmdOut) verboseOutResultForStep(idx int, sr *StepResult, path string, nest int) {
	indent := strings.Repeat("        ", nest)
	desc := ""
	if sr.Desc != "" {
		desc = fmt.Sprintf("%s ", sr.Desc)
	}
	switch {
	case sr.Err != nil:
		// fail
		lineformat := indent + "        %s\n"
		_, _ = fmt.Fprintf(d.out, "%s    --- %s(%s) ... %s\n%s", indent, desc, sr.Key, red("fail"), red(SprintMultilinef(lineformat, "Failure/Error: %s", strings.TrimRight(sr.Err.Error(), "\n"))))
		if sr.IncludedRunResult != nil {
			_, _ = fmt.Fprintf(d.out, "%s        === %s (%s)\n", indent, sr.IncludedRunResult.Desc, sr.IncludedRunResult.Path)
			sr.IncludedRunResult.included = false
			d.verboseOutResult(sr.IncludedRunResult, nest+1)
			return
		}
		b, err := readFile(path)
		if err != nil {
			return
		}
		picked, err := pickStepYAML(string(b), idx)
		if err != nil {
			return
		}
		_, _ = fmt.Fprintf(d.out, "%s        Failure step (%s):\n", indent, path)
		_, _ = fmt.Fprint(d.out, SprintMultilinef(lineformat, "%v", picked))
		_, _ = fmt.Fprintln(d.out, "")
	case sr.Skipped:
		// skip
		_, _ = fmt.Fprintf(d.out, "%s    --- %s(%s) ... %s\n", indent, desc, sr.Key, yellow("skip"))
		if sr.IncludedRunResult != nil {
			_, _ = fmt.Fprintf(d.out, "%s        === %s (%s)\n", indent, sr.IncludedRunResult.Desc, sr.IncludedRunResult.Path)
			sr.IncludedRunResult.included = false
			d.verboseOutResult(sr.IncludedRunResult, nest+1)
			return
		}
	default:
		// ok
		_, _ = fmt.Fprintf(d.out, "%s    --- %s(%s) ... %s\n", indent, desc, sr.Key, green("ok"))
		if sr.IncludedRunResult != nil {
			_, _ = fmt.Fprintf(d.out, "%s        === %s (%s)\n", indent, sr.IncludedRunResult.Desc, sr.IncludedRunResult.Path)
			sr.IncludedRunResult.included = false
			d.verboseOutResult(sr.IncludedRunResult, nest+1)
			return
		}
	}
}
