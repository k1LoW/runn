package runn

import (
	"context"
	"fmt"
	"os"
	"strconv"
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

const bpSep = ":"

type breakpoint struct {
	runbookID string
	stepKey   string
}

// dbg is runn debugger.
type dbg struct {
	enable      bool
	showPrompt  bool
	quit        bool
	breakpoints []breakpoint
	pp          *pp.PrettyPrinter
}

func newDBG(enable bool) *dbg {
	return &dbg{
		enable:     enable,
		showPrompt: true,
		pp:         pp.New(),
	}
}

func (d *dbg) attach(ctx context.Context, s *step) error {
	prpt := "> "

	if d.quit {
		s.parent.skipped = true
		return errStepSkiped
	}
	if !d.enable {
		return nil
	}

	if s != nil {
		id := s.parent.ID()
		stepKey := s.key
		stepIdx := strconv.Itoa(s.idx)
		// check breakpoints
		for _, bp := range d.breakpoints {
			if !strings.HasPrefix(id, bp.runbookID) {
				continue
			}
			if bp.stepKey != stepKey && bp.stepKey != stepIdx {
				continue
			}
			d.showPrompt = true
		}
		prpt = fmt.Sprintf("%s[%s]> ", id[:7], s.key)
	}

	if !d.showPrompt {
		return nil
	}
	d.showPrompt = false

L:
	for {
		in := prompt.Input(
			prompt.WithPrefix(prpt),
		)
		switch {
		case contains([]string{dbgCommandNext, dbgCommandNextShort}, in):
			// next
			d.showPrompt = true
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
			param := strings.TrimPrefix(strings.TrimPrefix(in, fmt.Sprintf("%s ", dbgCommandBreak)), fmt.Sprintf("%s ", dbgCommandBreakShort))
			splitted := strings.Split(param, bpSep)
			bp := breakpoint{}
			if splitted[0] != "" {
				bp.runbookID = splitted[0]
			} else {
				bp.runbookID = s.parent.ID()
			}
			if len(splitted) > 1 && splitted[1] != "" {
				bp.stepKey = splitted[1]
			} else {
				bp.stepKey = "0"
			}
			d.breakpoints = append(d.breakpoints, bp)
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
