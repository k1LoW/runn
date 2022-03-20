package runn

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/antonmedv/expr"
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

const rep = "-----"

var opReplacer = strings.NewReplacer("==", rep, "!=", rep, "<", rep, ">", rep, "<=", rep, ">=", rep, " not ", rep, "!", rep, " and ", rep, "&&", rep, " or ", rep, "||", rep)

func (rnr *testRunner) Run(ctx context.Context, cond string) error {
	store := rnr.operator.store.toMap()
	t := buildTree(cond, store)
	rnr.operator.Debugln("-----START TEST CONDITION-----")
	rnr.operator.Debugf("%s", t)
	rnr.operator.Debugln("-----END TEST CONDITION-----")
	tf, err := expr.Eval(fmt.Sprintf("(%s) == true", cond), store)
	if err != nil {
		return err
	}
	rnr.operator.record(nil)
	if !tf.(bool) {
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
	splitted := strings.Split(opReplacer.Replace(cond), rep)
	for _, p := range splitted {
		s := strings.Trim(p, " ")
		v, err := expr.Eval(s, store)
		if err != nil {
			tree.AddBranch(fmt.Sprintf("%s => ?", s))
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
