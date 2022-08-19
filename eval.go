package runn

import (
	"fmt"

	"github.com/antonmedv/expr"
)

func evalCond(cond string, store map[string]interface{}) (bool, error) {
	r, err := expr.Eval(cond, store)
	if err != nil {
		return false, err
	}
	str, ok := r.(string)
	if ok {
		switch str {
		case "true":
			return true, err
		case "false":
			return false, err
		}
	}
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
	c, ok := r.(int)
	if !ok {
		return 0, fmt.Errorf("invalid count: %v", count)
	}
	return c, nil
}
