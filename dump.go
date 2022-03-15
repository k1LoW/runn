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
	store := rnr.operator.store.toMap()
	v, err := expr.Eval(cond, store)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	out := string(b)
	_, _ = fmt.Fprintf(rnr.out, "%s\n", out)
	rnr.operator.record(map[string]interface{}{
		"out": out,
	})
	return nil
}
