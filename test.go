package runn

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	store := map[string]interface{}{
		"steps": rnr.operator.store.steps,
		"vars":  rnr.operator.store.vars,
	}
	if rnr.operator.debug {
		_, _ = fmt.Fprintln(os.Stderr, "-----START TEST CONDITION-----")
		tree := treeprint.New()
		tree.SetValue(cond)
		splitted := strings.Split(opReplacer.Replace(cond), rep)
		for _, p := range splitted {
			s := strings.Trim(p, " ")
			v, err := expr.Eval(s, store)
			if err != nil {
				tree.AddBranch(fmt.Sprintf("(%s) = ?", s))
				continue
			}
			b, err := json.Marshal(v)
			if err != nil {
				tree.AddBranch(fmt.Sprintf("(%s) = ?", s))
				continue
			}
			tree.AddBranch(fmt.Sprintf("(%s) = %s", s, string(b)))
		}
		_, _ = fmt.Fprint(os.Stderr, tree.String())
		_, _ = fmt.Fprintln(os.Stderr, "-----END TEST CONDITION-----")
	}
	tf, err := expr.Eval(fmt.Sprintf("(%s) == true", cond), store)
	if err != nil {
		return err
	}
	rnr.operator.store.steps = append(rnr.operator.store.steps, nil)
	if !tf.(bool) {
		return fmt.Errorf("(%s) is false", cond)
	}
	return nil
}
