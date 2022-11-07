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
	Typ     CDPArgType
	Key     string
	Example string
}

type CDPFnArgs []CDPFnArg

type CDPFn struct {
	Desc    string
	Fn      interface{}
	Args    CDPFnArgs
	Aliases []string
}

var CDPFnMap = map[string]CDPFn{
	"navigate": {
		Desc: "Navigate the current frame to `url` page.",
		Fn:   chromedp.Navigate,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "url", "https://pkg.go.dev/time"},
		},
	},
	"click": {
		Desc: "Send a mouse click event to the first element node matching the selector (`sel`).",
		Fn:   chromedp.Click,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "nav > div > a"},
		},
	},
	"doubleClick": {
		Desc: "Send a mouse double click event to the first element node matching the selector (`sel`).",
		Fn:   chromedp.DoubleClick,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "nav > div > li"},
		},
	},
	"sendKeys": {
		Desc: "Send keys (`value`) to the first element node matching the selector (`sel`).",
		Fn:   chromedp.SendKeys,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "input[name=username]"},
			{CDPArgTypeArg, "value", "k1lowxb@gmail.com"},
		},
	},
	"submit": {
		Desc: "Submit the parent form of the first element node matching the selector (`sel`).",
		Fn:   chromedp.Submit,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "form.login"},
		},
	},
	"wait": {
		Desc: "Wait for the specified `time`.",
		Fn: func(d string) chromedp.Action {
			return &waitAction{d: d}
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "time", "10sec"},
		},
		Aliases: []string{"sleep"},
	},
	"waitReady": {
		Desc: "Wait until the element matching the selector (`sel`) is ready.",
		Fn:   chromedp.WaitReady,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "body > footer"},
		},
	},
	"waitVisible": {
		Desc: "Wait until the element matching the selector (`sel`) is visible.",
		Fn:   chromedp.WaitVisible,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "body > footer"},
		},
	},
	"text": {
		Desc: "Get the visible text of the first element node matching the selector (`sel`).",
		Fn:   chromedp.Text,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "text", "Install the latest version of Go"},
		},
		Aliases: []string{"getText"},
	},
	"textContent": {
		Desc: "Get the text content of the first element node matching the selector (`sel`).",
		Fn:   chromedp.TextContent,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "text", "Install the latest version of Go"},
		},
		Aliases: []string{"getTextContent"},
	},
	"innerHTML": {
		Desc: "Get the inner html of the first element node matching the selector (`sel`).",
		Fn:   chromedp.InnerHTML,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "html", "Install the latest version of Go"},
		},
		Aliases: []string{"getInnerHTML"},
	},
	"outerHTML": {
		Desc: "Get the outer html of the first element node matching the selector (`sel`).",
		Fn:   chromedp.OuterHTML,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "html", "<h1>Install the latest version of Go</h1>"},
		},
		Aliases: []string{"getOuterHTML"},
	},
	"value": {
		Desc: "Get the Javascript value field of the first element node matching the selector (`sel`).",
		Fn:   chromedp.Value,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "input[name=address]"},
			{CDPArgTypeRes, "value", "Fukuoka"},
		},
		Aliases: []string{"getValue"},
	},
	"title": {
		Desc: "Get the document `title`.",
		Fn:   chromedp.Title,
		Args: CDPFnArgs{
			{CDPArgTypeRes, "title", "GitHub"},
		},
		Aliases: []string{"getTitle"},
	},
	"location": {
		Desc: "Get the document location.",
		Fn:   chromedp.Location,
		Args: CDPFnArgs{
			{CDPArgTypeRes, "url", "https://github.com"},
		},
		Aliases: []string{"getLocation"},
	},
	"attributes": {
		Desc: "Get the element attributes for the first element node matching the selector (`sel`).",
		Fn:   chromedp.Attributes,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "attrs", `{"class": "sr-only"}`},
		},
		Aliases: []string{"getAttributes", "attrs", "getAttrs"},
	},
	"evaluate": {
		Desc: "Evaluate the Javascript expression (`expr`).",
		Fn: func(expr string) chromedp.Action {
			// ignore the return value of 'evaluate'
			return chromedp.Evaluate(expr, nil)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "expr", `document.querySelector("h1").textContent = "hello"`},
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

func (args CDPFnArgs) ArgArgs() CDPFnArgs {
	res := CDPFnArgs{}
	for _, arg := range args {
		if arg.Typ == CDPArgTypeArg {
			res = append(res, arg)
		}
	}
	return res
}

func (args CDPFnArgs) ResArgs() CDPFnArgs {
	res := CDPFnArgs{}
	for _, arg := range args {
		if arg.Typ == CDPArgTypeRes {
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
