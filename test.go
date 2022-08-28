package runn

import (
	"context"
	"fmt"
	"strings"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/file"
	"github.com/antonmedv/expr/parser/lexer"
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
	source := file.NewSource(cond)
	tokens, _ := lexer.Lex(source)
	values := []string{}
	var (
		keep         string
		bracketStart bool
	)
	for i, t := range tokens {
		switch t.Kind {
		case lexer.EOF, lexer.Operator:
			continue
		case lexer.Bracket:
			if strings.ContainsAny(t.Value, "([{") {
				bracketStart = true
				keep = fmt.Sprintf("%s%s", keep, t.Value)
			} else if strings.ContainsAny(t.Value, ")]}") {
				bracketStart = false
				if i < len(tokens)-1 && tokens[i+1].Kind == lexer.Bracket {
					keep = fmt.Sprintf("%s%s", keep, t.Value)
					continue
				}
				if i < len(tokens)-1 && tokens[i+1].Is(lexer.Operator, ".") {
					keep = fmt.Sprintf("%s%s.", keep, t.Value)
					continue
				}
				values = append(values, fmt.Sprintf("%s%s", keep, t.Value))
				keep = ""
			}
		case lexer.Identifier:
			if i < len(tokens)-1 && tokens[i+1].Is(lexer.Operator, ".") {
				keep = fmt.Sprintf("%s%s.", keep, t.Value)
			} else if i < len(tokens)-1 && tokens[i+1].Kind == lexer.Bracket {
				keep = fmt.Sprintf("%s%s", keep, t.Value)
			} else {
				values = append(values, fmt.Sprintf("%s%s", keep, t.Value))
				keep = ""
			}
		case lexer.String:
			v := t.Value
			// lazy
			if strings.Contains(cond, fmt.Sprintf(`"%s"`, t.Value)) {
				v = fmt.Sprintf(`"%s"`, t.Value)
			} else if strings.Contains(cond, fmt.Sprintf(`'%s'`, t.Value)) {
				v = fmt.Sprintf(`'%s'`, t.Value)
			}
			if bracketStart {
				keep = fmt.Sprintf("%s%s", keep, v)
				continue
			}
			values = append(values, v)
		default:
			if bracketStart {
				keep = fmt.Sprintf("%s%s", keep, t.Value)
				continue
			}
			values = append(values, t.Value)
		}
	}
	return values
}
