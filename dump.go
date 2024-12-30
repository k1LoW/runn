package runn

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/runn/internal/expr"
	"github.com/k1LoW/runn/internal/store"
)

const dumpRunnerKey = "dump"

type dumpRunner struct{}

type dumpRequest struct {
	expr                   string
	out                    string
	disableTrailingNewline bool
	disableMaskingSecrets  bool
}

func newDumpRunner() *dumpRunner {
	return &dumpRunner{}
}

func (rnr *dumpRunner) Run(ctx context.Context, s *step, first bool) error {
	r := s.dumpRequest
	o := s.parent
	var out io.Writer
	sm := o.store.ToMap()
	sm[store.RootKeyIncluded] = o.included
	if first {
		if !s.deferred {
			sm[store.RootKeyPrevious] = o.store.Latest()
		}
	} else {
		if !s.deferred {
			sm[store.RootKeyPrevious] = o.store.Previous()
		}
		sm[store.RootKeyCurrent] = o.store.Latest()
	}
	if r.out == "" {
		if r.disableMaskingSecrets {
			out = o.stdout.Unwrap()
		} else {
			out = o.stdout
		}
	} else {
		p, err := expr.EvalExpand(r.out, sm)
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
			if err := donegroup.Cleanup(ctx, func() error {
				return f.Close()
			}); err != nil {
				return err
			}
			if r.disableMaskingSecrets {
				out = f
			} else {
				out = o.maskRule.NewWriter(f)
			}
		default:
			return fmt.Errorf("invalid dump out: %v", pp)
		}
	}
	v, err := expr.Eval(r.expr, sm)
	if err != nil {
		return err
	}
	if err := rnr.run(ctx, out, v, r.disableTrailingNewline, s, first); err != nil {
		return fmt.Errorf("failed to run dump: %w", err)
	}
	return nil
}

func (rnr *dumpRunner) run(_ context.Context, out io.Writer, v any, disableNL bool, s *step, first bool) error {
	o := s.parent
	switch vv := v.(type) {
	case string:
		if _, err := fmt.Fprint(out, vv); err != nil {
			return err
		}
	case []byte:
		// e.g. screenshot on CDP
		if _, err := out.Write(vv); err != nil {
			return err
		}
	default:
		if reflect.ValueOf(v).Kind() == reflect.Func {
			if _, err := fmt.Fprint(out, store.FuncValue); err != nil {
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
	if !disableNL {
		if _, err := fmt.Fprint(out, "\n"); err != nil {
			return err
		}
	}
	if first {
		o.record(s.idx, nil)
	}
	return nil
}
