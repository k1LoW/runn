package runn

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/file"
	"github.com/expr-lang/expr/parser"
	"github.com/expr-lang/expr/parser/lexer"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
	"github.com/xlab/treeprint"
)

const (
	delimStart = "{{"
	delimEnd   = "}}"
)

var alphaRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)

func Eval(e string, store any) (any, error) {
	v, err := expr.Eval(trimComment(e), store)
	if err != nil {
		return nil, fmt.Errorf("eval error: %w", err)
	}
	return v, nil
}

// EvalAny evaluate any type. but, EvalAny do not evaluate map key.
func EvalAny(e any, store any) (any, error) {
	switch v := e.(type) {
	case string:
		return Eval(v, store)
	case map[string]any:
		evaluated := make(map[string]any, len(v))
		for k, vv := range v {
			ev, err := EvalAny(vv, store)
			if err != nil {
				return nil, err
			}
			evaluated[k] = ev
		}
		return evaluated, nil
	case []any:
		evaluated := make([]any, len(v))
		for i, vv := range v {
			ev, err := EvalAny(vv, store)
			if err != nil {
				return nil, err
			}
			evaluated[i] = ev
		}
		return evaluated, nil
	default:
		return v, nil
	}
}

func EvalCond(cond string, store any) (bool, error) {
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

func EvalCount(count string, store any) (int, error) {
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
func EvalExpand(in, store any) (any, error) {
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

func buildTree(cond string, store any) (string, error) {
	if cond == "" {
		return "", nil
	}
	cond = trimComment(cond)
	tree := treeprint.New()
	tree.SetValue(fmt.Sprintf("%s\n│", cond))
	vs, err := values(cond)
	if err != nil {
		return "", err
	}
	for _, p := range vs {
		s := strings.Trim(p, " ")
		// string literal
		if strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`) {
			s = strings.Replace(s, "\n", "\\n", -1)
		}
		v, err := Eval(s, store)
		if err != nil {
			tree.AddBranch(fmt.Sprintf("%s => ?", s))
			continue
		}
		if vv, ok := v.(string); ok {
			tree.AddBranch(fmt.Sprintf(`%s => "%s"`, s, vv)) //nostyle:useq
			continue
		}
		b, err := json.Marshal(v)
		if err != nil {
			tree.AddBranch(fmt.Sprintf("%s => ?", s))
			continue
		}
		tree.AddBranch(fmt.Sprintf("%s => %s", s, string(b)))
	}
	return tree.String(), nil
}

func trimComment(cond string) string {
	const commentToken = "#"
	var b strings.Builder
	for _, l := range strings.Split(cond, "\n") {
		if strings.HasPrefix(l, commentToken) {
			continue
		}
		s := file.NewSource(l)
		tokens, err := lexer.Lex(s)
		if err != nil {
			b.WriteString(l)
			b.WriteByte('\n')
			continue
		}

		ccol := -1
		inClosure := false
	L:
		for _, t := range tokens {
			switch {
			case t.Kind == lexer.Bracket && t.Value == "{":
				inClosure = true
			case t.Kind == lexer.Bracket && t.Value == "}":
				inClosure = false
			case t.Kind == lexer.Operator && t.Value == commentToken && !inClosure:
				ccol = t.Column
				break L
			}
		}
		if ccol > 0 {
			b.WriteString(l[:ccol])
			b.WriteByte('\n')
			continue
		}

		b.WriteString(l)
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

func values(cond string) ([]string, error) {
	t, err := parser.Parse(cond)
	if err != nil {
		return nil, err
	}
	values := nodeValues(t.Node)
	return values, nil
}

func nodeValues(n ast.Node) []string {
	values := make([]string, 0, 2)
	switch v := n.(type) {
	case *ast.BinaryNode:
		values = append(values, nodeValues(v.Left)...)
		values = append(values, nodeValues(v.Right)...)
	case *ast.BoolNode:
		values = append(values, fmt.Sprintf("%v", v.Value))
	case *ast.StringNode:
		values = append(values, fmt.Sprintf("%q", v.Value))
	case *ast.IntegerNode:
		values = append(values, fmt.Sprintf("%d", v.Value))
	case *ast.FloatNode:
		values = append(values, fmt.Sprintf("%v", v.Value))
	case *ast.ArrayNode:
		values = append(values, arrayNode(v))
	case *ast.MapNode:
		values = append(values, mapNode(v))
	case *ast.IdentifierNode:
		values = append(values, v.Value)
	case *ast.NilNode:
		values = append(values, fmt.Sprintf(`%v`, nil))
	case *ast.BuiltinNode:
		values = append(values, builtinNode(v)...)
	case *ast.MemberNode:
		values = append(values, memberNode(v))
	case *ast.UnaryNode:
		values = append(values, unaryNode(v))
	case *ast.CallNode:
		values = append(values, callNode(v)...)
	case *ast.ClosureNode:
		values = append(values, closureNode(v))
	case *ast.PointerNode:
		values = append(values, "#")
	}
	return values
}

func nodeValue(n ast.Node) string {
	ns := nodeValues(n)
	if len(ns) != 1 {
		return ""
	}
	return ns[0]
}

func arrayNode(a *ast.ArrayNode) string {
	var builder strings.Builder
	builder.WriteByte('[')
	for i := 0; i < len(a.Nodes); i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(nodeValue(a.Nodes[i]))
	}
	builder.WriteByte(']')
	return builder.String()
}

func mapNode(m *ast.MapNode) string {
	var builder strings.Builder
	builder.WriteByte('{')
	for i := 0; i < len(m.Pairs); i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		pair, ok := m.Pairs[i].(*ast.PairNode)
		if !ok {
			continue
		}
		builder.WriteString(strings.Trim(nodeValue(pair.Key), `"`))
		builder.WriteString(": ")
		builder.WriteString(nodeValue(pair.Value))
	}
	builder.WriteByte('}')
	return builder.String()
}

func memberNode(m *ast.MemberNode) string {
	n := nodeValue(m.Node)
	switch v := m.Property.(type) {
	case *ast.StringNode:
		if alphaRe.MatchString(v.Value) {
			return fmt.Sprintf("%s.%s", n, v.Value)
		}
		return fmt.Sprintf(`%s[%q]`, n, v.Value)
	case *ast.IntegerNode:
		return fmt.Sprintf("%s[%d]", n, v.Value)
	case *ast.IdentifierNode:
		return fmt.Sprintf("%s[%s]", n, v.Value)
	default:
		return fmt.Sprintf("%s.%s", n, nodeValue(v))
	}
}

func unaryNode(u *ast.UnaryNode) string {
	return u.Operator + nodeValue(u.Node)
}

func callNode(c *ast.CallNode) []string {
	var builder strings.Builder
	builder.WriteString(nodeValue(c.Callee))
	builder.WriteByte('(')
	for i := 0; i < len(c.Arguments); i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(nodeValues(c.Arguments[i])[0])
	}
	builder.WriteByte(')')
	values := []string{builder.String()}
	args := make([]string, len(c.Arguments))
	argValues := make([]string, 0, len(c.Arguments))
	for i := range c.Arguments {
		vs := nodeValues(c.Arguments[i])
		args[i] = vs[0]
		argValues = append(argValues, vs[1:]...)
	}
	return append(append(values, args...), argValues...)
}

func builtinNode(b *ast.BuiltinNode) []string {
	var builder strings.Builder
	builder.WriteString(b.Name)
	builder.WriteByte('(')
	for i := 0; i < len(b.Arguments); i++ {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(nodeValue(b.Arguments[i]))
	}
	builder.WriteByte(')')
	values := []string{builder.String()}
	args := make([]string, len(b.Arguments))
	argValues := make([]string, 0, len(b.Arguments))
	for i := range b.Arguments {
		v := b.Arguments[i]
		if _, ok := v.(*ast.ClosureNode); !ok {
			argValues = append(argValues, nodeValue(v))
		}
		args[i] = nodeValue(v)
	}
	return append(append(values, args...), argValues...)
}

func closureNode(c *ast.ClosureNode) string {
	return fmt.Sprintf("{ %s }", nodeValue(c.Node))
}
