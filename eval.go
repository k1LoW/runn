package runn

import (
	"fmt"
	"strconv"

	"github.com/antonmedv/expr"
)

func evalCond(cond string, store map[string]interface{}) (bool, error) {
	tf, err := expr.Eval(fmt.Sprintf("(%s) == true", cond), store)
	if err != nil {
		return false, err
	}
	return tf.(bool), nil
}

func evalCount(count string, store map[string]interface{}) (int, error) {
	r, err := expr.Eval(count, store)
	if err != nil {
		return 0, err
	}
	var c int
	switch v := r.(type) {
	case string:
		c, err = strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid count: evaluated %s, but got %T(%v): %w", count, r, r, err)
		}
	case int64:
		c = int(v)
	case float64:
		c = int(v)
	case int:
		c = v
	default:
		return 0, fmt.Errorf("invalid count: evaluated %s, but got %T(%v)", count, r, r)
	}
	return c, nil
}
