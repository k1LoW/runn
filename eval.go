package runn

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/file"
	"github.com/antonmedv/expr/parser"
	"github.com/antonmedv/expr/parser/lexer"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/expand"
	"github.com/xlab/treeprint"
)

const (
	delimStart = "{{"
	delimEnd   = "}}"
)

var (
	expandRe = regexp.MustCompile(fmt.Sprintf(`"?%s\s*([^}]+)\s*%s"?`, delimStart, delimEnd))
	numberRe = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?$`)
)

func eval(e string, store map[string]interface{}) (interface{}, error) {
	return expr.Eval(trimComment(e), store)
}

func evalCond(cond string, store map[string]interface{}) (bool, error) {
	v, err := eval(cond, store)
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

func evalCount(count string, store map[string]interface{}) (int, error) {
	r, err := eval(count, store)
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

func evalExpand(in interface{}, store map[string]interface{}) (interface{}, error) {
	b, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	var reperr error
	replacefunc := func(in string) string {
		if !strings.Contains(in, delimStart) {
			return in
		}
		matches := expandRe.FindAllStringSubmatch(in, -1)
		oldnew := []string{}
		for _, m := range matches {
			o, err := eval(m[1], store)
			if err != nil {
				reperr = err
				return ""
			}
			var s string
			switch v := o.(type) {
			case string:
				// Stringify only one expression.
				if strings.TrimSpace(in) == m[0] && numberRe.MatchString(v) {
					s = fmt.Sprintf("'%s'", v)
				} else {
					s = v
				}
			case int64:
				s = strconv.Itoa(int(v))
			case uint64:
				s = strconv.Itoa(int(v))
			case float64:
				s = strconv.FormatFloat(v, 'f', -1, 64)
			case int:
				s = strconv.Itoa(v)
			case bool:
				s = strconv.FormatBool(v)
			case map[string]interface{}, []interface{}:
				bytes, err := json.Marshal(v)
				if err != nil {
					reperr = fmt.Errorf("json.Marshal error: %w", err)
				} else {
					s = string(bytes)
				}
			default:
				reperr = fmt.Errorf("invalid format: evaluated %s, but got %T(%v)", m[1], o, o)
				return ""
			}
			oldnew = append(oldnew, m[0], s)
		}
		rep := strings.NewReplacer(oldnew...)
		return rep.Replace(in)
	}
	e := expand.ReplaceYAML(string(b), replacefunc, true)
	if reperr != nil {
		return nil, reperr
	}
	var out interface{}
	if err := yaml.Unmarshal([]byte(e), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func buildTree(cond string, store map[string]interface{}) (string, error) {
	if cond == "" {
		return "", nil
	}
	cond = trimComment(cond)
	tree := treeprint.New()
	tree.SetValue(cond)
	vs, err := values(cond)
	if err != nil {
		return "", err
	}
	for _, p := range vs {
		s := strings.Trim(p, " ")
		v, err := eval(s, store)
		if err != nil {
			tree.AddBranch(fmt.Sprintf("%s => ?", s))
			continue
		}
		if vv, ok := v.(string); ok {
			tree.AddBranch(fmt.Sprintf(`%s => "%s"`, s, vv))
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
	trimed := []string{}
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
		for _, t := range tokens {
			switch {
			case t.Kind == lexer.Bracket && t.Value == "{":
				inClosure = true
			case t.Kind == lexer.Bracket && t.Value == "}":
				inClosure = false
			case t.Kind == lexer.Operator && t.Value == commentToken && inClosure == false:
				ccol = t.Column
				break
			}
		}
		if ccol > 0 {
			trimed = append(trimed, strings.TrimSuffix(l[:ccol], " "))
			continue
		}

		trimed = append(trimed, l)
	}
	return strings.Join(trimed, "\n")
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
	values := []string{}
	switch v := n.(type) {
	case *ast.BinaryNode:
		values = append(values, nodeValues(v.Left)...)
		values = append(values, nodeValues(v.Right)...)
	case *ast.BoolNode:
		values = append(values, fmt.Sprintf(`%v`, v.Value))
	case *ast.StringNode:
		values = append(values, fmt.Sprintf(`"%s"`, v.Value))
	case *ast.IntegerNode:
		values = append(values, fmt.Sprintf(`%d`, v.Value))
	case *ast.FloatNode:
		values = append(values, fmt.Sprintf(`%v`, v.Value))
	case *ast.ArrayNode:
		values = append(values, arrayNode(v))
	case *ast.MapNode:
		values = append(values, mapNode(v))
	case *ast.IdentifierNode:
		values = append(values, v.Value)
	case *ast.PropertyNode:
		values = append(values, propertyNode(v))
	case *ast.IndexNode:
		values = append(values, indexNode(v))
	case *ast.FunctionNode:
		values = append(values, functionNode(v)...)
	case *ast.NilNode:
		values = append(values, fmt.Sprintf(`%v`, nil))
	case *ast.BuiltinNode:
		values = append(values, builtinNode(v)...)
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
	elems := []string{}
	for _, e := range a.Nodes {
		elems = append(elems, nodeValue(e))
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
}

func mapNode(m *ast.MapNode) string {
	kvs := []string{}
	for _, p := range m.Pairs {
		switch v := p.(type) {
		case *ast.PairNode:
			kvs = append(kvs, fmt.Sprintf("%s: %s", strings.Trim(nodeValue(v.Key), `"`), nodeValue(v.Value)))
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(kvs, ", "))
}

func propertyNode(p *ast.PropertyNode) string {
	return fmt.Sprintf("%s.%s", nodeValue(p.Node), p.Property)
}

func indexNode(i *ast.IndexNode) string {
	n := nodeValue(i.Node)
	switch v := i.Index.(type) {
	case *ast.StringNode:
		return fmt.Sprintf(`%s["%s"]`, n, v.Value)
	case *ast.IntegerNode:
		return fmt.Sprintf(`%s[%d]`, n, v.Value)
	case *ast.IdentifierNode:
		return fmt.Sprintf(`%s[%s]`, n, v.Value)
	default:
		return ""
	}
}

func functionNode(f *ast.FunctionNode) []string {
	args := []string{}
	for _, a := range f.Arguments {
		args = append(args, nodeValue(a))
	}
	values := []string{fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ", "))}
	return append(values, args...)
}

func builtinNode(b *ast.BuiltinNode) []string {
	args := []string{}
	for _, a := range b.Arguments {
		args = append(args, nodeValue(a))
	}
	values := []string{fmt.Sprintf("%s(%s)", b.Name, strings.Join(args, ", "))}
	return append(values, args...)
}
