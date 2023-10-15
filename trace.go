package runn

type trace struct {
	RunID string `json:"id"`
}

func NewTrace(o *operator) trace {
	return trace{
		RunID: o.id,
	}
}
