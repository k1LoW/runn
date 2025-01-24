package runn

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/k1LoW/donegroup"
)

const cdpNewKey = "new"

const (
	cdpTimeoutByStep = 60 * time.Second
	cdpWindowWidth   = 1920
	cdpWindowHeight  = 1080
)

type cdpRunner struct {
	name          string
	ctx           context.Context //nostyle:contexts
	cancel        context.CancelFunc
	store         map[string]any
	opts          []chromedp.ExecAllocatorOption
	timeoutByStep time.Duration
	mu            sync.Mutex
	// operatorID - The id of the operator for which the runner is defined.
	operatorID string
}

type CDPActions []CDPAction

type CDPAction struct {
	Fn   string
	Args map[string]any
}

func newCDPRunner(name, remote string) (*cdpRunner, error) {
	if remote != cdpNewKey {
		return nil, errors.New("remote connect mode is planned, but not yet implemented")
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(cdpWindowWidth, cdpWindowHeight),
	)

	if os.Getenv("RUNN_DISABLE_HEADLESS") != "" {
		opts = append(opts,
			chromedp.Flag("headless", false),
			chromedp.Flag("hide-scrollbars", false),
			chromedp.Flag("mute-audio", false),
		)
	}

	return &cdpRunner{
		name:          name,
		store:         map[string]any{},
		opts:          opts,
		timeoutByStep: cdpTimeoutByStep,
	}, nil
}

func (rnr *cdpRunner) Close() error {
	rnr.mu.Lock()
	defer rnr.mu.Unlock()
	if rnr.cancel == nil {
		return nil
	}
	rnr.cancel()
	rnr.ctx = nil
	rnr.cancel = nil
	return nil
}

func (rnr *cdpRunner) Renew() error {
	if err := rnr.Close(); err != nil {
		return err
	}
	rnr.store = map[string]any{}
	return nil
}

func (rnr *cdpRunner) Run(ctx context.Context, s *step) error {
	o := s.parent
	cas, err := parseCDPActions(s.cdpActions, s, o.expandBeforeRecord)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}
	if err := rnr.run(ctx, cas, s); err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}
	return nil
}

func (rnr *cdpRunner) run(ctx context.Context, cas CDPActions, s *step) error {
	o := s.parent
	if rnr.ctx == nil {
		allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), rnr.opts...)
		ctxx, _ := chromedp.NewContext(allocCtx, chromedp.WithDebugf(func(format string, a ...any) {
			fmt.Println("chromedp debug:", fmt.Sprintf(format, a...))
		}))
		rnr.ctx = ctxx
		rnr.cancel = cancel
		// Merge run() function context and runner (chrome) context
		context.AfterFunc(ctxx, func() {
			_ = rnr.Close()
		})
		if err := donegroup.Cleanup(ctx, func() error {
			// In the case of Reused runners, leave the cleanup to the main cleanup
			if o.id != rnr.operatorID {
				return nil
			}
			return rnr.Renew()
		}); err != nil {
			return err
		}
	}
	o.capturers.captureCDPStart(rnr.name)
	defer o.capturers.captureCDPEnd(rnr.name)

	// Set a timeout (cdpTimeoutByStep) for each step because Chrome operations may get stuck depending on the actions: specified.
	called := atomic.Bool{}
	defer func() {
		called.Store(true)
	}()
	timer := time.NewTimer(rnr.timeoutByStep)
	go func() {
		<-timer.C
		if !called.Load() {
			rnr.Close()
		}
	}()

	before := []chromedp.Action{
		chromedp.EmulateViewport(cdpWindowWidth, cdpWindowHeight),
	}
	if err := chromedp.Run(rnr.ctx, before...); err != nil {
		return err
	}
	for i, ca := range cas {
		o.capturers.captureCDPAction(ca)
		k, fn, err := findCDPFn(ca.Fn)
		if err != nil {
			return fmt.Errorf("actions[%d] error: %w", i, err)
		}
		if k == "latestTab" {
			infos, err := chromedp.Targets(rnr.ctx)
			if err != nil {
				return err
			}
			latestCtx, _ := chromedp.NewContext(rnr.ctx, chromedp.WithTargetID(infos[0].TargetID))
			rnr.ctx = latestCtx
			continue
		}
		as, err := rnr.evalAction(ca, s)
		if err != nil {
			return fmt.Errorf("actions[%d] error: %w", i, err)
		}
		if err := chromedp.Run(rnr.ctx, as...); err != nil {
			return fmt.Errorf("actions[%d] error: %w", i, err)
		}
		ras := fn.Args.ResArgs()
		if len(ras) > 0 {
			// capture
			res := map[string]any{}
			for _, arg := range ras {
				v := rnr.store[arg.Key]
				switch vv := v.(type) {
				case *string:
					res[arg.Key] = *vv
				case *map[string]string:
					res[arg.Key] = *vv
				case *[]byte:
					res[arg.Key] = *vv
				default:
					res[arg.Key] = vv
				}
			}
			o.capturers.captureCDPResponse(ca, res)
		}
	}

	// record
	r := map[string]any{}
	for k, v := range rnr.store {
		switch vv := v.(type) {
		case *string:
			r[k] = *vv
		case *map[string]string:
			r[k] = *vv
		case *[]byte:
			r[k] = *vv
		default:
			r[k] = vv
		}
	}
	o.record(s.idx, r)

	rnr.store = map[string]any{} // clear

	return nil
}

func (rnr *cdpRunner) evalAction(ca CDPAction, s *step) ([]chromedp.Action, error) {
	o := s.parent
	_, fn, err := findCDPFn(ca.Fn)
	if err != nil {
		return nil, err
	}

	// path resolution for setUploadFile.path
	if ca.Fn == "setUploadFile" {
		p, ok := ca.Args["path"]
		if !ok {
			return nil, fmt.Errorf("invalid action: %v: arg %q not found", ca, "path")
		}
		pp, ok := p.(string)
		if !ok {
			return nil, fmt.Errorf("invalid action: %v", ca)
		}
		ca.Args["path"], err = fp(pp, o.root)
		if err != nil {
			return nil, fmt.Errorf("invalid action: %v: %w", ca, err)
		}
	}

	fv := reflect.ValueOf(fn.Fn)
	var vs []reflect.Value
	for i, a := range fn.Args {
		switch a.Typ {
		case CDPArgTypeArg:
			v, ok := ca.Args[a.Key]
			if !ok {
				return nil, fmt.Errorf("invalid action: %v: arg %q not found", ca, a.Key)
			}
			if v == nil {
				return nil, fmt.Errorf("invalid action arg: %s.%s = %v", ca.Fn, a.Key, v)
			}
			vs = append(vs, reflect.ValueOf(v))
		case CDPArgTypeRes:
			k := a.Key
			switch reflect.TypeOf(fn.Fn).In(i).Elem().Kind() {
			case reflect.String:
				var v string
				rnr.store[k] = &v
				vs = append(vs, reflect.ValueOf(&v))
			case reflect.Map:
				// e.g. attributes
				v := map[string]string{}
				rnr.store[k] = &v
				vs = append(vs, reflect.ValueOf(&v))
			case reflect.Slice:
				var v []byte
				rnr.store[k] = &v
				vs = append(vs, reflect.ValueOf(&v))
			default:
				return nil, fmt.Errorf("invalid action: %v", ca)
			}
		default:
			return nil, fmt.Errorf("invalid action: %v", ca)
		}
	}
	res := fv.Call(vs)
	a, ok := res[0].Interface().(chromedp.Action)
	if ok {
		return []chromedp.Action{a}, nil
	}
	as, ok := res[0].Interface().([]chromedp.Action)
	if ok {
		return as, nil
	}
	return nil, fmt.Errorf("invalid action: %v", ca)
}
