package runn

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/k1LoW/duration"
)

type CDPArgType string

const (
	CDPArgTypeArg CDPArgType = "arg"
	CDPArgTypeRes CDPArgType = "res"
)

type CDPFnArg struct {
	typ CDPArgType
	key string
}

type CDPFnArgs []CDPFnArg

type CDPFn struct {
	Fn      interface{}
	Args    CDPFnArgs
	Aliases []string
}

var CDPFnMap = map[string]CDPFn{
	"navigate": {
		Fn: chromedp.Navigate,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "url"},
		},
	},
	"click": {
		Fn: chromedp.Click,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
		},
	},
	"doubleClick": {
		Fn: chromedp.DoubleClick,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
		},
	},
	"sendKeys": {
		Fn: chromedp.SendKeys,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeArg, "value"},
		},
	},
	"submit": {
		Fn: chromedp.Submit,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
		},
	},
	"wait": {
		Fn: func(d string) chromedp.Action {
			return &waitAction{d: d}
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "time"},
		},
		Aliases: []string{"sleep"},
	},
	"waitReady": {
		Fn: chromedp.WaitReady,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
		},
	},
	"waitVisible": {
		Fn: chromedp.WaitVisible,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
		},
	},
	"text": {
		Fn: chromedp.Text,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "text"},
		},
		Aliases: []string{"getText"},
	},
	"textContent": {
		Fn: chromedp.TextContent,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "text"},
		},
		Aliases: []string{"getTextContent"},
	},
	"innerHTML": {
		Fn: chromedp.InnerHTML,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "html"},
		},
		Aliases: []string{"getInnerHTML"},
	},
	"outerHTML": {
		Fn: chromedp.OuterHTML,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "html"},
		},
		Aliases: []string{"getOuterHTML"},
	},
	"value": {
		Fn: chromedp.Value,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "value"},
		},
		Aliases: []string{"getValue"},
	},
	"title": {
		Fn: chromedp.Title,
		Args: CDPFnArgs{
			{CDPArgTypeRes, "title"},
		},
		Aliases: []string{"getTitle"},
	},
	"location": {
		Fn: chromedp.Location,
		Args: CDPFnArgs{
			{CDPArgTypeRes, "url"},
		},
		Aliases: []string{"getLocation"},
	},
	"attributes": {
		Fn: chromedp.Attributes,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "attrs"},
		},
		Aliases: []string{"getAttributes", "attrs", "getAttrs"},
	},
	"evaluate": {
		Fn: func(expr string) chromedp.Action {
			return chromedp.Evaluate(expr, nil)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "expr"},
		},
		Aliases: []string{"eval"},
	},
}

func findCDPFn(k string) (CDPFn, error) {
	fn, ok := CDPFnMap[k]
	if ok {
		return fn, nil
	}
	for _, fn := range CDPFnMap {
		for _, a := range fn.Aliases {
			if a == k {
				return fn, nil
			}
		}
	}
	return CDPFn{}, fmt.Errorf("not found function: %s", k)
}

func (args CDPFnArgs) resArgs() CDPFnArgs {
	res := CDPFnArgs{}
	for _, arg := range args {
		if arg.typ == CDPArgTypeRes {
			res = append(res, arg)
		}
	}
	return res
}

var _ chromedp.Action = (*waitAction)(nil)

type waitAction struct {
	d string
}

func (w *waitAction) Do(ctx context.Context) error {
	d, err := duration.Parse(w.d)
	if err != nil {
		return err
	}
	time.Sleep(d)
	return nil
}
