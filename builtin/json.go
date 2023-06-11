package builtin

import "encoding/json"

type JSON struct{}

func NewJSON() *JSON {
	return &JSON{}
}

func (j *JSON) Encode(in interface{}) interface{} {
	b, err := json.Marshal(in)
	if err != nil {
		return nil
	}
	return string(b)
}

func (j *JSON) Decode(in string) interface{} {
	var out interface{}
	if err := json.Unmarshal([]byte(in), &out); err != nil {
		return nil
	}
	return out
}
