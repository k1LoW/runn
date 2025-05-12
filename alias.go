package runn

import (
	"github.com/k1LoW/runn/internal/expr"
	"github.com/k1LoW/runn/internal/exprtrace"
	"github.com/k1LoW/runn/internal/scope"
)

const (
	AllowReadParent = scope.AllowReadParent
	AllowReadRemote = scope.AllowReadRemote
	AllowRunExec    = scope.AllowRunExec //nostyle:repetition
)

// EvalWithTrace evaluates an expression with tracing and returns the evaluation result with trace information.
// This is useful for debugging expressions and understanding how they are evaluated.
func EvalWithTrace(e string, store exprtrace.EvalEnv) (*exprtrace.EvalResult, error) {
	return expr.EvalWithTrace(e, store)
}

// Eval evaluates an expression using the provided environment store and returns the result.
func Eval(e string, store exprtrace.EvalEnv) (any, error) {
	return expr.Eval(e, store)
}

// EvalAny evaluate any type. but, EvalAny do not evaluate map key.
// EvalAny evaluates any type of value, recursively evaluating expressions in strings.
// Note that map keys are not evaluated.
func EvalAny(e any, store exprtrace.EvalEnv) (any, error) {
	return expr.EvalAny(e, store)
}

// EvalCond evaluates a condition expression and returns a boolean result.
func EvalCond(cond string, store exprtrace.EvalEnv) (bool, error) {
	return expr.EvalCond(cond, store)
}

// EvalCount evaluates an expression that should result in an integer value.
// This is typically used for loop counts or other numeric expressions.
func EvalCount(count string, store exprtrace.EvalEnv) (int, error) {
	return expr.EvalCount(count, store)
}

// EvalExpand evaluates `in` and expand `{{ }}` in `in` using `store`.
// EvalExpand evaluates an input value and expands any expressions in {{ }} using the provided store.
// This is used for template-like string interpolation within configuration values.
func EvalExpand(in any, store exprtrace.EvalEnv) (any, error) {
	return expr.EvalExpand(in, store)
}
