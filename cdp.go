package runn

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/chromedp/chromedp"
)

const cdpNewKey = "new"

type CDPArgType string

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

type cdpActions []cdpAction

type cdpAction struct {
	fn   string
	args map[string]interface{}
}

func newCDPRunner(name, remote string) (*cdpRunner, error) {
	if remote != cdpNewKey {
		return nil, errors.New("remote connect mode is planned, but not yet implemented")
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(cdpWindowWidth, cdpWindowHeight),
	)
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

func (rnr *cdpRunner) Run(_ context.Context, cas cdpActions) error {
	as := []chromedp.Action{
		chromedp.EmulateViewport(cdpWindowWidth, cdpWindowHeight),
	}
	for _, ca := range cas {
		a, err := rnr.evalAction(ca)
		if err != nil {
			return err
		}
		as = append(as, a)
	}
	if err := chromedp.Run(rnr.ctx, as...); err != nil {
		return err
	}

	// record
	r := map[string]interface{}{}
	for k, v := range rnr.store {
		switch vv := v.(type) {
		case *string:
			r[k] = *vv
		default:
			r[k] = vv
		}
	}
	rnr.operator.record(r)

	rnr.store = map[string]interface{}{} // clear

	return nil
}

func (rnr *cdpRunner) evalAction(ca cdpAction) (chromedp.Action, error) {
	fn, ok := CDPFnMap[ca.fn]
	if !ok {
		return nil, fmt.Errorf("invalid action: %v", ca)
	}
	fv := reflect.ValueOf(fn.Fn)
	vs := []reflect.Value{}
	for i, a := range fn.Args {
		switch a.typ {
		case CDPArgTypeArg:
			v, ok := ca.args[a.key]
			if !ok {
				return nil, fmt.Errorf("invalid action: %v", ca)
			}
			vs = append(vs, reflect.ValueOf(v))
		case CDPArgTypeRes, CDPArgTypeHiddenRes:
			k := a.key

			switch reflect.TypeOf(fn.Fn).In(i).Kind() {
			case reflect.Interface:
				// evaluate
				var v interface{}
				rnr.store[k] = &v
				vs = append(vs, reflect.ValueOf(&v))
			default:
				switch reflect.TypeOf(fn.Fn).In(i).Elem().Kind() {
				case reflect.String:
					var v string
					rnr.store[k] = &v
					vs = append(vs, reflect.ValueOf(&v))
				default:
					return nil, fmt.Errorf("invalid action: %v", ca)
				}
			}
		}
	}
	res := fv.Call(vs)
	return res[0].Interface().(chromedp.Action), nil
}
