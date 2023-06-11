package builtin

func Compare(x, y any, ignoreKeys ...string) bool {
	d, err := diff(x, y, ignoreKeys...)
	if err != nil {
		return false
	}

	return d == ""
}
