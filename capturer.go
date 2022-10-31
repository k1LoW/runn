package runn

import (
	"net/http"

	"google.golang.org/grpc/metadata"
)

type Capturer interface {
	CaptureStart(ids IDs, bookPath, desc string)
	CaptureFailed(ids IDs, bookPath, desc string, err error)
	CaptureSkipped(ids IDs, bookPath, desc string)
	CaptureSuccess(ids IDs, bookPath, desc string)
	CaptureEnd(ids IDs, bookPath, desc string)

	CaptureHTTPRequest(name string, req *http.Request)
	CaptureHTTPResponse(name string, res *http.Response)

	CaptureGRPCStart(name string, typ GRPCType, service, method string)
	CaptureGRPCRequestHeaders(h map[string][]string)
	CaptureGRPCRequestMessage(m map[string]interface{})
	CaptureGRPCResponseStatus(status int)
	CaptureGRPCResponseHeaders(h map[string][]string)
	CaptureGRPCResponseMessage(m map[string]interface{})
	CaptureGRPCResponseTrailers(t map[string][]string)
	CaptureGRPCClientClose()
	CaptureGRPCEnd(name string, typ GRPCType, service, method string)

	CaptureDBStatement(name string, stmt string)
	CaptureDBResponse(name string, res *DBResponse)

	CaptureExecCommand(command string)
	CaptureExecStdin(stdin string)
	CaptureExecStdout(stdin string)
	CaptureExecStderr(stderr string)

	SetCurrentIDs(ids IDs)
	Errs() error
}

type capturers []Capturer

func (cs capturers) captureStart(ids IDs, bookPath, desc string) {
	for _, c := range cs {
		c.CaptureStart(ids, bookPath, desc)
	}
}

func (cs capturers) captureFailed(ids IDs, bookPath, desc string, err error) {
	for _, c := range cs {
		c.CaptureFailed(ids, bookPath, desc, err)
	}
}

func (cs capturers) captureSkipped(ids IDs, bookPath, desc string) {
	for _, c := range cs {
		c.CaptureSkipped(ids, bookPath, desc)
	}
}

func (cs capturers) captureSuccess(ids IDs, bookPath, desc string) {
	for _, c := range cs {
		c.CaptureSuccess(ids, bookPath, desc)
	}
}

func (cs capturers) captureEnd(ids IDs, bookPath, desc string) {
	for _, c := range cs {
		c.CaptureEnd(ids, bookPath, desc)
	}
}

func (cs capturers) captureHTTPRequest(name string, req *http.Request) {
	for _, c := range cs {
		c.CaptureHTTPRequest(name, req)
	}
}

func (cs capturers) captureHTTPResponse(name string, res *http.Response) {
	for _, c := range cs {
		c.CaptureHTTPResponse(name, res)
	}
}

func (cs capturers) captureGRPCStart(name string, typ GRPCType, service, method string) {
	for _, c := range cs {
		c.CaptureGRPCStart(name, typ, service, method)
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

func (cs capturers) captureGRPCClientClose() {
	for _, c := range cs {
		c.CaptureGRPCClientClose()
	}
}

func (cs capturers) captureGRPCEnd(name string, typ GRPCType, service, method string) {
	for _, c := range cs {
		c.CaptureGRPCEnd(name, typ, service, method)
	}
}

func (cs capturers) captureDBStatement(name string, stmt string) {
	for _, c := range cs {
		c.CaptureDBStatement(name, stmt)
	}
}

func (cs capturers) captureDBResponse(name string, res *DBResponse) {
	for _, c := range cs {
		c.CaptureDBResponse(name, res)
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

func (cs capturers) setCurrentIDs(ids IDs) {
	for _, c := range cs {
		c.SetCurrentIDs(ids)
	}
}
