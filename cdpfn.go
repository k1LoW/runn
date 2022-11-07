package runn

import "github.com/chromedp/chromedp"

const (
	CDPArgTypeArg       CDPArgType = "arg"
	CDPArgTypeRes       CDPArgType = "res"
	CDPArgTypeHiddenRes CDPArgType = "hidden"
)

type CDPFnArg struct {
	typ CDPArgType
	key string
}

type CDPFnArgs []CDPFnArg

type CDPFn struct {
	Fn   interface{}
	Args CDPFnArgs
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
	},
	"textContent": {
		Fn: chromedp.TextContent,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "text"},
		},
	},
	"innerHTML": {
		Fn: chromedp.InnerHTML,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "html"},
		},
	},
	"outerHTML": {
		Fn: chromedp.OuterHTML,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "html"},
		},
	},
	"value": {
		Fn: chromedp.Value,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel"},
			{CDPArgTypeRes, "value"},
		},
	},
	"title": {
		Fn: chromedp.Title,
		Args: CDPFnArgs{
			{CDPArgTypeRes, "title"},
		},
	},
	"location": {
		Fn: chromedp.Location,
		Args: CDPFnArgs{
			{CDPArgTypeRes, "url"},
		},
	},
	"evaluate": {
		Fn: chromedp.Evaluate,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "expr"},
			{CDPArgTypeHiddenRes, "res"},
		},
	},
}
