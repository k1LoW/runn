package runn

import (
	"net/http"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Capturer interface {
	CaptureStart(trs Trails, bookPath, desc string)
	CaptureResult(trs Trails, result *RunResult)
	CaptureEnd(trs Trails, bookPath, desc string)

	CaptureResultByStep(trs Trails, result *RunResult)

	CaptureHTTPRequest(name string, req *http.Request)
	CaptureHTTPResponse(name string, res *http.Response)

	CaptureGRPCStart(name string, typ GRPCType, service, method string)
	CaptureGRPCRequestHeaders(h map[string][]string)
	CaptureGRPCRequestMessage(m map[string]any)
	CaptureGRPCResponseStatus(*status.Status)
	CaptureGRPCResponseHeaders(h map[string][]string)
	CaptureGRPCResponseMessage(m map[string]any)
	CaptureGRPCResponseTrailers(t map[string][]string)
	CaptureGRPCClientClose()
	CaptureGRPCEnd(name string, typ GRPCType, service, method string)

	CaptureCDPStart(name string)
	CaptureCDPAction(a CDPAction)
	CaptureCDPResponse(a CDPAction, res map[string]any)
	CaptureCDPEnd(name string)

	CaptureSSHCommand(command string)
	CaptureSSHStdout(stdout string)
	CaptureSSHStderr(stderr string)

	CaptureDBStatement(name string, stmt string)
	CaptureDBResponse(name string, res *DBResponse)

	CaptureExecCommand(command, shell string)
	CaptureExecStdin(stdin string)
	CaptureExecStdout(stdout string)
	CaptureExecStderr(stderr string)

	SetCurrentTrails(trs Trails)
	Errs() error
}

type capturers []Capturer

func (cs capturers) captureStart(trs Trails, bookPath, desc string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureStart(trs, bookPath, desc)
	}
}

func (cs capturers) captureResult(trs Trails, result *RunResult) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureResult(trs, result)
	}
}

func (cs capturers) captureEnd(trs Trails, bookPath, desc string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureEnd(trs, bookPath, desc)
	}
}

func (cs capturers) captureResultByStep(trs Trails, result *RunResult) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureResultByStep(trs, result)
	}
}

func (cs capturers) captureHTTPRequest(name string, req *http.Request) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureHTTPRequest(name, req)
	}
}

func (cs capturers) captureHTTPResponse(name string, res *http.Response) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureHTTPResponse(name, res)
	}
}

func (cs capturers) captureGRPCStart(name string, typ GRPCType, service, method string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCStart(name, typ, service, method)
	}
}
func (cs capturers) captureGRPCRequestHeaders(h metadata.MD) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCRequestHeaders(h)
	}
}

func (cs capturers) captureGRPCRequestMessage(m map[string]any) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCRequestMessage(m)
	}
}

func (cs capturers) captureGRPCResponseStatus(s *status.Status) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCResponseStatus(s)
	}
}

func (cs capturers) captureGRPCResponseHeaders(h metadata.MD) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCResponseHeaders(h)
	}
}

func (cs capturers) captureGRPCResponseMessage(m map[string]any) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCResponseMessage(m)
	}
}

func (cs capturers) captureGRPCResponseTrailers(t metadata.MD) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCResponseTrailers(t)
	}
}

func (cs capturers) captureGRPCClientClose() { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCClientClose()
	}
}

func (cs capturers) captureGRPCEnd(name string, typ GRPCType, service, method string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureGRPCEnd(name, typ, service, method)
	}
}

func (cs capturers) captureCDPStart(name string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureCDPStart(name)
	}
}

func (cs capturers) captureCDPAction(a CDPAction) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureCDPAction(a)
	}
}

func (cs capturers) captureCDPResponse(a CDPAction, res map[string]any) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureCDPResponse(a, res)
	}
}

func (cs capturers) captureCDPEnd(name string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureCDPEnd(name)
	}
}

func (cs capturers) captureSSHCommand(command string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureSSHCommand(command)
	}
}

func (cs capturers) captureSSHStdout(stdout string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureSSHStdout(stdout)
	}
}

func (cs capturers) captureSSHStderr(stderr string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureSSHStderr(stderr)
	}
}

func (cs capturers) captureDBStatement(name string, stmt string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureDBStatement(name, stmt)
	}
}

func (cs capturers) captureDBResponse(name string, res *DBResponse) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureDBResponse(name, res)
	}
}

func (cs capturers) captureExecCommand(command, shell string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureExecCommand(command, shell)
	}
}

func (cs capturers) captureExecStdin(stdin string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureExecStdin(stdin)
	}
}

func (cs capturers) captureExecStdout(stdout string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureExecStdout(stdout)
	}
}

func (cs capturers) captureExecStderr(stderr string) { //nostyle:recvtype
	for _, c := range cs {
		c.CaptureExecStderr(stderr)
	}
}

func (cs capturers) setCurrentTrails(trs Trails) { //nostyle:recvtype
	for _, c := range cs {
		c.SetCurrentTrails(trs)
	}
}
