package builtin

import (
	"fmt"

	"github.com/samber/lo"
)

func Omit(x any, keys ...string) any {
	d, err := omit(x, keys...)
	if err != nil {
		panic(err)
	}

	return d
}

func omit(x any, keys ...string) (any, error) {
	if t, ok := x.(map[string]any); ok {
		return lo.OmitByKeys(t, keys), nil
	} else {
		return nil, fmt.Errorf("unsupported type: %T", x)
	}
}
