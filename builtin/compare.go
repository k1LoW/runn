package builtin

func Compare(x, y any, ignorePaths ...string) bool {
	d, err := diff(x, y, ignorePaths...)
	if err != nil {
		return false
	}

	return d == ""
}
