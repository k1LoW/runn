package builtin

import "encoding/json"

type JSON struct{}

func NewJSON() *JSON {
	return &JSON{}
}

func (j *JSON) Encode(in any) any {
	b, err := json.Marshal(in)
	if err != nil {
		return nil
	}
	return string(b)
}

func (j *JSON) Decode(in string) any {
	var out any
	if err := json.Unmarshal([]byte(in), &out); err != nil {
		return nil
	}
	return out
}
