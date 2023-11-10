package runn

type trace struct {
	RunID string `json:"id"`
}

func newTrace(s *step) trace {
	return trace{
		RunID: s.runbookID(),
	}
}
