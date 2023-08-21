package builtin

type JSON struct{}

func NewJSON() *JSON {
	return &JSON{}
}

func (j *JSON) Encode(in any) any {
	panic("json.Encode() is deprecated, use toJSON() instead")
}

func (j *JSON) Decode(in string) any {
	panic("json.Decode() is deprecated, use fromJSON() instead")
}
