package runn

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/goccy/go-json"
)

const dumpRunnerKey = "dump"

type dumpRunner struct{}

type dumpRequest struct {
	expr string
	out  string
}

func newDumpRunner() (*dumpRunner, error) {
	return &dumpRunner{}, nil
}

func (rnr *dumpRunner) Run(ctx context.Context, s *step, first bool) error {
	o := s.parent
	r := s.dumpRequest
	var out io.Writer
	store := o.store.toMap()
	store[storeIncludedKey] = o.included
	if first {
		store[storePreviousKey] = o.store.latest()
	} else {
		store[storePreviousKey] = o.store.previous()
		store[storeCurrentKey] = o.store.latest()
	}
	if r.out == "" {
		out = o.stdout
	} else {
		p, err := EvalExpand(r.out, store)
		if err != nil {
			return err
		}
		switch pp := p.(type) {
		case string:
			if !filepath.IsAbs(pp) {
				pp = filepath.Join(filepath.Dir(o.bookPath), pp)
			}
			f, err := os.Create(pp)
			if err != nil {
				return err
			}
			out = f
		default:
			return fmt.Errorf("invalid dump out: %v", pp)
		}
	}
	v, err := Eval(r.expr, store)
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
		if reflect.ValueOf(v).Kind() == reflect.Func {
			if _, err := fmt.Fprint(out, storeFuncValue); err != nil {
				return err
			}
		} else {
			b, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return err
			}
			if _, err := fmt.Fprint(out, string(b)); err != nil {
				return err
			}
		}
	}
	if r.out == "" {
		if _, err := fmt.Fprint(out, "\n"); err != nil {
			return err
		}
	}
	if first {
		o.record(nil)
	}
	return nil
}
