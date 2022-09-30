package runn

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
)

func evalCond(cond string, store map[string]interface{}) (bool, error) {
	tf, err := expr.Eval(fmt.Sprintf("(%s) == true", trimComment(cond)), store)
	if err != nil {
		return false, err
	}
	return tf.(bool), nil
}

func evalCount(count string, store map[string]interface{}) (int, error) {
	r, err := expr.Eval(trimComment(count), store)
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

func trimComment(cond string) string {
	const commentToken = "#"
	trimed := []string{}
	for _, l := range strings.Split(cond, "\n") {
		if strings.HasPrefix(strings.Trim(l, " "), commentToken) {
			continue
		}
		trimed = append(trimed, l)
	}
	return strings.Join(trimed, "\n")
}
