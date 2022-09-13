package runn

import (
	"net/http"

	"google.golang.org/grpc/metadata"
)

type Capturer interface {
	CaptureStart(ids []string, bookPath string)
	CaptureEnd(ids []string, bookPath string)

	CaptureHTTPRequest(req *http.Request)
	CaptureHTTPResponse(res *http.Response)

	CaptureGRPCStart(service, method string)
	CaptureGRPCRequestHeaders(h map[string][]string)
	CaptureGRPCRequestMessage(m map[string]interface{})
	CaptureGRPCResponseStatus(status int)
	CaptureGRPCResponseHeaders(h map[string][]string)
	CaptureGRPCResponseMessage(m map[string]interface{})
	CaptureGRPCResponseTrailers(t map[string][]string)
	CaptureGRPCEnd(service, method string)

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

func (cs capturers) captureStart(ids []string, bookPath string) {
	for _, c := range cs {
		c.CaptureStart(ids, bookPath)
	}
}

func (cs capturers) captureEnd(ids []string, bookPath string) {
	for _, c := range cs {
		c.CaptureEnd(ids, bookPath)
	}
}

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

func (cs capturers) captureGRPCStart(service, method string) {
	for _, c := range cs {
		c.CaptureGRPCStart(service, method)
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

func (cs capturers) captureGRPCEnd(service, method string) {
	for _, c := range cs {
		c.CaptureGRPCEnd(service, method)
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
