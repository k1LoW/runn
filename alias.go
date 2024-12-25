package runn

import (
	"github.com/k1LoW/runn/internal/eval"
	"github.com/k1LoW/runn/internal/exprtrace"
)

func EvalWithTrace(e string, store exprtrace.EvalEnv) (*exprtrace.EvalResult, error) {
	return eval.EvalWithTrace(e, store)
}

func Eval(e string, store exprtrace.EvalEnv) (any, error) {
	return eval.Eval(e, store)
}

// EvalAny evaluate any type. but, EvalAny do not evaluate map key.
func EvalAny(e any, store exprtrace.EvalEnv) (any, error) {
	return eval.EvalAny(e, store)
}

func EvalCond(cond string, store exprtrace.EvalEnv) (bool, error) {
	return eval.EvalCond(cond, store)
}

func EvalCount(count string, store exprtrace.EvalEnv) (int, error) {
	return eval.EvalCount(count, store)
}

// EvalExpand evaluates `in` and expand `{{ }}` in `in` using `store`.
func EvalExpand(in any, store exprtrace.EvalEnv) (any, error) {
	return eval.EvalExpand(in, store)
}
