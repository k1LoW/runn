package runn

import (
	"context"
	"fmt"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/parser"
	"github.com/goccy/go-json"
	"github.com/xlab/treeprint"
)

const testRunnerKey = "test"

type testRunner struct {
	operator *operator
}

func newTestRunner(o *operator) (*testRunner, error) {
	return &testRunner{
		operator: o,
	}, nil
}

func (rnr *testRunner) Run(ctx context.Context, cond string, runned bool) error {
	store := rnr.operator.store.toMap()
	if runned {
		store[storeCurrentKey] = rnr.operator.store.latest()
	}
	t := buildTree(cond, store)
	rnr.operator.Debugln("-----START TEST CONDITION-----")
	rnr.operator.Debugf("%s", t)
	rnr.operator.Debugln("-----END TEST CONDITION-----")
	tf, err := evalCond(cond, store)
	if err != nil {
		return err
	}
	if !tf {
		return fmt.Errorf("(%s) is not true\n%s", cond, t)
	}
	return nil
}

func buildTree(cond string, store map[string]interface{}) string {
	if cond == "" {
		return ""
	}
	tree := treeprint.New()
	tree.SetValue(cond)
	for _, p := range values(cond) {
		s := strings.Trim(p, " ")
		v, err := expr.Eval(s, store)
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
	return tree.String()
}

func values(cond string) []string {
	t, _ := parser.Parse(cond)
	values := nodeValues(t.Node)
	return values
}

func nodeValues(n ast.Node) []string {
	values := []string{}
	switch v := n.(type) {
	case *ast.BinaryNode:
		values = append(values, nodeValues(v.Left)...)
		values = append(values, nodeValues(v.Right)...)
	case *ast.StringNode:
		values = append(values, fmt.Sprintf(`"%s"`, v.Value))
	case *ast.IntegerNode:
		values = append(values, fmt.Sprintf(`%d`, v.Value))
	case *ast.FloatNode:
		values = append(values, fmt.Sprintf(`%v`, v.Value))
	case *ast.ArrayNode:
		values = append(values, arrayNode(v))
	case *ast.IdentifierNode:
		values = append(values, v.Value)
	case *ast.PropertyNode:
		values = append(values, propertyNode(v))
	case *ast.IndexNode:
		values = append(values, indexNode(v))
	case *ast.FunctionNode:
		values = append(values, functionNode(v)...)
	}
	return values
}

func arrayNode(a *ast.ArrayNode) string {
	elems := []string{}
	for _, e := range a.Nodes {
		n := nodeValues(e)
		if len(n) != 1 {
			return ""
		}
		elems = append(elems, n[0])
	}
	return fmt.Sprintf("[%s]", strings.Join(elems, ", "))
}

func propertyNode(p *ast.PropertyNode) string {
	n := nodeValues(p.Node)
	if len(n) != 1 {
		return ""
	}
	return fmt.Sprintf("%s.%s", n[0], p.Property)
}

func indexNode(i *ast.IndexNode) string {
	n := nodeValues(i.Node)
	if len(n) != 1 {
		return ""
	}
	switch v := i.Index.(type) {
	case *ast.StringNode:
		return fmt.Sprintf(`%s["%s"]`, n[0], v.Value)
	case *ast.IntegerNode:
		return fmt.Sprintf(`%s[%d]`, n[0], v.Value)
	default:
		return ""
	}
}

func functionNode(f *ast.FunctionNode) []string {
	args := []string{}
	for _, a := range f.Arguments {
		n := nodeValues(a)
		if len(n) != 1 {
			return nil
		}
		args = append(args, n[0])
	}
	values := []string{fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ", "))}
	return append(values, args...)
}
