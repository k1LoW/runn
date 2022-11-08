package runn

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-json"
)

const dumpRunnerKey = "dump"

type dumpRunner struct {
	operator *operator
	out      io.Writer
}

type dumpRequest struct {
	cond string
	out  string
}

func newDumpRunner(o *operator) (*dumpRunner, error) {
	return &dumpRunner{
		operator: o,
		out:      os.Stdout,
	}, nil
}

func (rnr *dumpRunner) Run(ctx context.Context, r *dumpRequest) error {
	var out io.Writer
	if r.out == "" {
		out = rnr.out
	}
	store := rnr.operator.store.toNormalizedMap()
	store[storePreviousKey] = rnr.operator.store.previous()
	store[storeCurrentKey] = rnr.operator.store.latest()
	v, err := eval(r.cond, store)
	if err != nil {
		return err
	}
	switch vv := v.(type) {
	case string:
		_, _ = fmt.Fprint(out, vv)
	case []byte:
		_, _ = fmt.Fprint(out, vv)
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		_, _ = fmt.Fprint(out, string(b))
	}
	if r.out == "" {
		_, _ = fmt.Fprint(out, "\n")
	}
	return nil
}
