package eval

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/file"
	"github.com/expr-lang/expr/parser/lexer"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
	"github.com/k1LoW/runn/builtin"
	"github.com/k1LoW/runn/internal/deprecation"
	"github.com/k1LoW/runn/internal/exprtrace"
	"github.com/xlab/treeprint"
)

const (
	delimStart = "{{"
	delimEnd   = "}}"
)

var baseTreePrinterOptions = []exprtrace.TreePrinterOption{
	exprtrace.WithOnCallNodeHook(onPrintTraceCallNode),
}

var evalMu sync.Mutex //FIXME: remove global mutex

func EvalWithTrace(e string, store exprtrace.EvalEnv) (*exprtrace.EvalResult, error) {
	e = trimDeprecatedComment(e)
	var result *exprtrace.EvalResult

	trace := exprtrace.NewStore()
	tracer := exprtrace.NewTracer(trace, store)
	evalMu.Lock()
	program, err := expr.Compile(e, tracer.Patches()...)
	evalMu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("eval error: %w", err)
	}
	env := tracer.InstallTracerFunctions(store)
	out, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("eval error: %w", err)
	}
	result = &exprtrace.EvalResult{
		Output:             out,
		Trace:              trace,
		Source:             e,
		Env:                env,
		TreePrinterOptions: baseTreePrinterOptions,
	}

	return result, nil
}

func Eval(e string, store exprtrace.EvalEnv) (any, error) {
	v, err := expr.Eval(trimDeprecatedComment(e), store)
	if err != nil {
		return nil, fmt.Errorf("eval error: %w", err)
	}
	return v, nil
}

// EvalAny evaluate any type. but, EvalAny do not evaluate map key.
func EvalAny(e any, store exprtrace.EvalEnv) (any, error) {
	switch v := e.(type) {
	case string:
		return Eval(v, store)
	case map[string]any:
		evaluated := map[string]any{}
		for k, vv := range v {
			ev, err := EvalAny(vv, store)
			if err != nil {
				return nil, err
			}
			evaluated[k] = ev
		}
		return evaluated, nil
	case []any:
		var evaluated []any
		for _, vv := range v {
			ev, err := EvalAny(vv, store)
			if err != nil {
				return nil, err
			}
			evaluated = append(evaluated, ev)
		}
		return evaluated, nil
	default:
		return v, nil
	}
}

func EvalCond(cond string, store exprtrace.EvalEnv) (bool, error) {
	v, err := Eval(cond, store)
	if err != nil {
		return false, err
	}
	switch vv := v.(type) {
	case bool:
		return vv, nil
	default:
		return false, nil
	}
}

func EvalCount(count string, store exprtrace.EvalEnv) (int, error) {
	r, err := Eval(count, store)
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
	case uint64:
		if v > math.MaxInt {
			return 0, fmt.Errorf("invalid count: evaluated %s, but got %T(%v): %w", count, r, r, err)
		}
		c = int(v) //nolint:gosec
	case float64:
		c = int(v)
	case int:
		c = v
	default:
		return 0, fmt.Errorf("invalid count: evaluated %s, but got %T(%v)", count, r, r)
	}
	return c, nil
}

// EvalExpand evaluates `in` and expand `{{ }}` in `in` using `store`.
func EvalExpand(in any, store exprtrace.EvalEnv) (any, error) {
	if s, ok := in.(string); ok {
		if !strings.Contains(s, delimStart) {
			// No need to expand
			return in, nil
		}
		if !strings.Contains(s, ":") {
			// Single value
			repFn := expand.ExprRepFn(delimStart, delimEnd, store)
			e, err := repFn(s)
			if err != nil {
				return nil, err
			}
			if strings.Contains(e, "\n") {
				// Multi line string literal
				return e, nil
			}
			// Single line string or number or bool or...
			var out any
			if err := yaml.Unmarshal([]byte(e), &out); err != nil {
				return nil, err
			}
			return out, nil
		}
	}
	// Expand using expand.ExprRepFn
	b, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	e, err := expand.ReplaceYAML(string(b), expand.ExprRepFn(delimStart, delimEnd, store), expand.ReplaceMapKey())
	if err != nil {
		return nil, err
	}
	var out any
	if err := yaml.Unmarshal([]byte(e), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func trimDeprecatedComment(cond string) string {
	const commentToken = "#"
	s := file.NewSource(cond)
	tokens, err := lexer.Lex(s)
	if err != nil {
		return cond
	}
	found := false
	inClosure := false
	inClosure2 := false
	for _, t := range tokens {
		switch {
		case t.Kind == lexer.Bracket && t.Value == "{":
			inClosure = true
		case t.Kind == lexer.Bracket && t.Value == "}":
			inClosure = false
		case t.Kind == lexer.Bracket && t.Value == "(":
			inClosure2 = true
		case t.Kind == lexer.Bracket && t.Value == ")":
			inClosure2 = false
		case t.Kind == lexer.Operator && t.Value == commentToken && !inClosure && !inClosure2:
			found = true
		}
	}
	if !found {
		return cond
	}

	// Deprecated comment token
	deprecation.AddWarning("Sharp comment", "`#` comment is deprecated. Use `//` instead.")

	var trimed []string
	for _, l := range strings.Split(cond, "\n") {
		if strings.HasPrefix(strings.Trim(l, " "), commentToken) {
			continue
		}
		s := file.NewSource(l)
		tokens, err := lexer.Lex(s)
		if err != nil {
			trimed = append(trimed, l)
			continue
		}

		ccol := -1
		inClosure := false
		inClosure2 := false
	L:
		for _, t := range tokens {
			switch {
			case t.Kind == lexer.Bracket && t.Value == "{":
				inClosure = true
			case t.Kind == lexer.Bracket && t.Value == "}":
				inClosure = false
			case t.Kind == lexer.Bracket && t.Value == "(":
				inClosure2 = true
			case t.Kind == lexer.Bracket && t.Value == ")":
				inClosure2 = false
			case t.Kind == lexer.Operator && t.Value == commentToken && !inClosure && !inClosure2:
				ccol = t.To - 1
				break L
			}
		}
		if ccol > 0 {
			trimed = append(trimed, strings.TrimSuffix(l[:ccol], " "))
			continue
		}

		trimed = append(trimed, l)
	}
	return strings.TrimRight(strings.Join(trimed, "\n"), "\n")
}

func onPrintTraceCallNode(tp exprtrace.TreePrinter, tree treeprint.Tree, callNode *ast.CallNode, callOutput any) {
	if callNode.Callee.String() == "compare" && len(callNode.Arguments) >= 2 && callOutput == false &&
		reflect.ValueOf(tp.EvalEnv()["compare"]).Pointer() == reflect.ValueOf(builtin.Compare).Pointer() {

		a := tp.PeekNodeEvalResultOutput(callNode.Arguments[0])
		b := tp.PeekNodeEvalResultOutput(callNode.Arguments[1])

		ignoreKeys := make([]string, 0, len(callNode.Arguments)-2)
		for i := range callNode.Arguments[2:] {
			if key, ok := tp.PeekNodeEvalResultOutput(callNode.Arguments[i+2]).(string); ok {
				ignoreKeys = append(ignoreKeys, key)
			}
		}

		diff, _ := builtin.Diff(a, b, ignoreKeys) //nostyle:handlerror

		// Normalize NBSP to SPACE
		diff = strings.Map(func(r rune) rune {
			if r == '\u00A0' {
				return ' '
			}
			return r
		}, diff)
		// ref.) https://github.com/google/go-cmp/issues/235

		tree.AddNode(fmt.Sprintf("(diff) => %s", strings.TrimSuffix(diff, "\n")))
	}
}
