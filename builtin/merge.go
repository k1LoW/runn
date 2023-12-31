package builtin

import (
	"fmt"

	"github.com/samber/lo"
)

func Merge(x ...any) any {
	d, err := merge(x...)
	if err != nil {
		panic(err)
	}

	return d
}

func merge(x ...any) (any, error) {
	y := make([]map[string]any, len(x))
	for _, t := range x {
		if t, ok := t.(map[string]any); ok {
			y = append(y, t)
		} else {
			return nil, fmt.Errorf("unsupported type: %T", x)
		}
	}
	return lo.Assign(y...), nil
}
