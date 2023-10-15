package runn

type trace struct {
	RunID string `json:"id"`
}

func NewTrace(s *step) trace {
	return trace{
		RunID: s.runbookIDFull(),
	}
}
