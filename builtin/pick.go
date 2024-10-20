package builtin

import (
	"fmt"

	"github.com/samber/lo"
)

func Pick(x any, keys ...string) (any, error) {
	if t, ok := x.(map[string]any); ok {
		return lo.PickByKeys(t, keys), nil
	} else {
		return nil, fmt.Errorf("unsupported type: %T", x)
	}
}
