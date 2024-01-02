package runn

const traceHeaderName = "X-Runn-Trace"

type trace struct {
	RunID string `json:"id"`
}

func newTrace(s *step) trace {
	return trace{
		RunID: s.runbookID(),
	}
}
