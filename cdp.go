package runn

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/chromedp/chromedp"
)

const cdpNewKey = "new"

const (
	cdpWindowWidth  = 1920
	cdpWindowHeight = 1080
)

type cdpRunner struct {
	name     string
	ctx      context.Context
	cancel   context.CancelFunc
	store    map[string]interface{}
	operator *operator
}

type CDPActions []CDPAction

type CDPAction struct {
	Fn   string
	Args map[string]interface{}
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

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx)
	return &cdpRunner{
		name:   name,
		ctx:    ctx,
		cancel: cancel,
		store:  map[string]interface{}{},
	}, nil
}

func (rnr *cdpRunner) Close() error {
	if rnr.cancel == nil {
		return nil
	}
	rnr.cancel()
	rnr.cancel = nil
	return nil
}

func (rnr *cdpRunner) Run(_ context.Context, cas CDPActions) error {
	rnr.operator.capturers.captureCDPStart(rnr.name)
	defer rnr.operator.capturers.captureCDPEnd(rnr.name)
	before := []chromedp.Action{
		chromedp.EmulateViewport(cdpWindowWidth, cdpWindowHeight),
	}
	if err := chromedp.Run(rnr.ctx, before...); err != nil {
		return err
	}
	for _, ca := range cas {
		rnr.operator.capturers.captureCDPAction(ca)
		as, err := rnr.evalAction(ca)
		if err != nil {
			return err
		}
		if err := chromedp.Run(rnr.ctx, as...); err != nil {
			return err
		}
		fn, err := findCDPFn(ca.Fn)
		if err != nil {
			return err
		}
		ras := fn.Args.ResArgs()
		if len(ras) > 0 {
			// capture
			res := map[string]interface{}{}
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
			rnr.operator.capturers.captureCDPResponse(ca, res)
		}
	}

	// record
	r := map[string]interface{}{}
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
	rnr.operator.record(r)

	rnr.store = map[string]interface{}{} // clear

	return nil
}

func (rnr *cdpRunner) evalAction(ca CDPAction) ([]chromedp.Action, error) {
	fn, err := findCDPFn(ca.Fn)
	if err != nil {
		return nil, err
	}
	fv := reflect.ValueOf(fn.Fn)
	vs := []reflect.Value{}
	for i, a := range fn.Args {
		switch a.Typ {
		case CDPArgTypeArg:
			v, ok := ca.Args[a.Key]
			if !ok {
				return nil, fmt.Errorf("invalid action: %v: arg '%s' not found", ca, a.Key)
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
				// ex. attributes
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
