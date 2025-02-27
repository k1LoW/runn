package runn

import (
	"github.com/k1LoW/runn/internal/expr"
	"github.com/k1LoW/runn/internal/exprtrace"
)

func EvalWithTrace(e string, store exprtrace.EvalEnv) (*exprtrace.EvalResult, error) {
	return expr.EvalWithTrace(e, store)
}

func Eval(e string, store exprtrace.EvalEnv) (any, error) {
	return expr.Eval(e, store)
}

// EvalAny evaluate any type. but, EvalAny do not evaluate map key.
func EvalAny(e any, store exprtrace.EvalEnv) (any, error) {
	return expr.EvalAny(e, store)
}

func EvalCond(cond string, store exprtrace.EvalEnv) (bool, error) {
	return expr.EvalCond(cond, store)
}

func EvalCount(count string, store exprtrace.EvalEnv) (int, error) {
	return expr.EvalCount(count, store)
}

// EvalExpand evaluates `in` and expand `{{ }}` in `in` using `store`.
func EvalExpand(in any, store exprtrace.EvalEnv) (any, error) {
	return expr.EvalExpand(in, store)
}
