package builtin

func Compare(x, y interface{}, ignoreKeys ...string) bool {
	d, err := diff(x, y, ignoreKeys...)
	if err != nil {
		return false
	}

	return d == ""
}
