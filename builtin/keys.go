package builtin

import (
	"fmt"

	"github.com/samber/lo"
)

func Keys(x any) any {
	d, err := keys(x)
	if err != nil {
		panic(err)
	}

	return d
}

func keys(x any) (any, error) {
	switch x := x.(type) {
	case map[string]any:
		return lo.Keys(x), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", x)
	}
}
