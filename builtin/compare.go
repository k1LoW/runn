package builtin

func Compare(x, y any, ignores ...any) (bool, error) {
	d, err := Diff(x, y, ignores...)
	if err != nil {
		return false, err
	}

	return (d == ""), nil
}
