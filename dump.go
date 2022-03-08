package runn

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/antonmedv/expr"
	"github.com/goccy/go-json"
)

const dumpRunnerKey = "dump"

type dumpRunner struct {
	operator *operator
	out      io.Writer
}

func newDumpRunner(o *operator) (*dumpRunner, error) {
	return &dumpRunner{
		operator: o,
		out:      os.Stderr,
	}, nil
}

func (rnr *dumpRunner) Run(ctx context.Context, cond string) error {
	store := map[string]interface{}{
		"steps": rnr.operator.store.steps,
		"vars":  rnr.operator.store.vars,
	}
	v, err := expr.Eval(cond, store)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(rnr.out, "%s\n", string(b))
	return nil
}
