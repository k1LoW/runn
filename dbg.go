package runn

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/elk-language/go-prompt"
	"github.com/k0kubun/pp/v3"
)

const (
	dbgCommandNext          = "next"
	dbgCommandNextShort     = "n"
	dbgCommandPrint         = "print"
	dbgCommandPrintShort    = "p"
	dbgCommandQuit          = "quit"
	dbgCommandQuitShort     = "q"
	dbgCommandBreak         = "break"
	dbgCommandBreakShort    = "b"
	dbgCommandContinue      = "continue"
	dbgCommandContinueShort = "c"
)

// dbg is runn debugger.
type dbg struct {
	enable bool
	step   bool
	quit   bool
	pp     *pp.PrettyPrinter
}

func newDBG(enable bool) *dbg {
	return &dbg{
		enable: enable,
		step:   true,
		pp:     pp.New(),
	}
}

func (d *dbg) attach(ctx context.Context, s *step) error {
	if d.quit {
		s.parent.skipped = true
		return errStepSkiped
	}
	if !d.enable {
		return nil
	}
	if !d.step {
		return nil
	}
	d.step = false

	prpt := "> "
	if s != nil {
		id := s.parent.ID()[:7]
		prpt = fmt.Sprintf("%s[%s]> ", id, s.key)
	}

L:
	for {
		in := prompt.Input(
			prompt.WithPrefix(prpt),
		)
		switch {
		case contains([]string{dbgCommandNext, dbgCommandNextShort}, in):
			// next
			d.step = true
			break L
		case contains([]string{dbgCommandContinue, dbgCommandContinueShort}, in):
			// continue
			break L
		case contains([]string{dbgCommandQuit, dbgCommandQuitShort}, in):
			// quit
			d.quit = true
			s.parent.skipped = true
			return errStepSkiped
		case strings.HasPrefix(in, fmt.Sprintf("%s ", dbgCommandPrint)) || strings.HasPrefix(in, fmt.Sprintf("%s ", dbgCommandPrintShort)):
			// print
			param := strings.TrimPrefix(strings.TrimPrefix(in, fmt.Sprintf("%s ", dbgCommandPrint)), fmt.Sprintf("%s ", dbgCommandPrintShort))
			store := s.parent.store.toMap()
			store[storeRootKeyIncluded] = s.parent.included
			store[storeRootKeyPrevious] = s.parent.store.latest()
			e, err := Eval(param, store)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				continue
			}
			d.pp.Println(e)
		case strings.HasPrefix(in, fmt.Sprintf("%s ", dbgCommandBreak)) || strings.HasPrefix(in, fmt.Sprintf("%s ", dbgCommandBreakShort)):
			// break
			_, _ = fmt.Fprintf(os.Stderr, "not implemented %s\n", in)
		default:
			_, _ = fmt.Fprintf(os.Stderr, "unknown command %s\n", in)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}
