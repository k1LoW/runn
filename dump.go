package runn

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
)

const dumpRunnerKey = "dump"

type dumpRunner struct {
	operator *operator
	out      io.Writer
}

type dumpRequest struct {
	expr string
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
	} else {
		p := r.out
		if !filepath.IsAbs(r.out) {
			p = filepath.Join(filepath.Dir(rnr.operator.bookPath), r.out)
		}
		f, err := os.Create(p)
		if err != nil {
			return err
		}
		out = f
	}
	store := rnr.operator.store.toNormalizedMap()
	store[storePreviousKey] = rnr.operator.store.previous()
	store[storeCurrentKey] = rnr.operator.store.latest()
	v, err := eval(r.expr, store)
	if err != nil {
		return err
	}
	switch vv := v.(type) {
	case string:
		if _, err := fmt.Fprint(out, vv); err != nil {
			return err
		}
	case []byte:
		// ex. screenshot on CDP
		if _, err := out.Write(vv); err != nil {
			return err
		}
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		if _, err := fmt.Fprint(out, string(b)); err != nil {
			return err
		}
	}
	if r.out == "" {
		if _, err := fmt.Fprint(out, "\n"); err != nil {
			return err
		}
	}
	return nil
}
