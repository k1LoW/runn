package builtin

func Compare(x, y any, ignorePaths ...string) (bool, error) {
	d, err := Diff(x, y, ignorePaths...)
	if err != nil {
		return false, err
	}

	return (d == ""), nil
}
