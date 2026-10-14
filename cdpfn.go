package runn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/domstorage"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
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
	Fn      any
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
	"tabTo": {
		Desc: "Change current frame to the tab with the specified `url`.",
		Fn: func() chromedp.Action[chromedp.Void] {
			// dummy
			return nil
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "url", "https://pkg.go.dev/time"},
		},
	},
	"click": {
		Desc: "Send a mouse click event to the first element node matching the selector (`sel`).",
		Fn:   chromedp.Click[string],
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "nav > div > a"},
		},
	},
	"doubleClick": {
		Desc: "Send a mouse double click event to the first element node matching the selector (`sel`).",
		Fn:   chromedp.DoubleClick[string],
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "nav > div > li"},
		},
	},
	"sendKeys": {
		Desc: "Send keys (`value`) to the first element node matching the selector (`sel`).",
		Fn:   chromedp.SendKeys[string],
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "input[name=username]"},
			{CDPArgTypeArg, "value", "k1lowxb@gmail.com"},
		},
	},
	"submit": {
		Desc: "Submit the parent form of the first element node matching the selector (`sel`).",
		Fn:   chromedp.Submit[string],
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "form.login"},
		},
	},
	"scroll": {
		Desc: "Scroll the window to the first element node matching the selector (`sel`).",
		Fn:   chromedp.ScrollIntoView[string],
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "body > footer"},
		},
		Aliases: []string{"scrollIntoView"},
	},
	"wait": {
		Desc: "Wait for the specified `time`.",
		Fn: func(d string) chromedp.Action[chromedp.Void] {
			return chromedp.Legacy(&waitAction{d: d})
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "time", "10sec"},
		},
		Aliases: []string{"sleep"},
	},
	"waitReady": {
		Desc: "Wait until the element matching the selector (`sel`) is ready.",
		Fn:   chromedp.WaitReady[string],
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "body > footer"},
		},
	},
	"waitVisible": {
		Desc: "Wait until the element matching the selector (`sel`) is visible.",
		Fn:   chromedp.WaitVisible[string],
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "body > footer"},
		},
	},
	"setUserAgent": {
		Desc: "Set the default User-Agent",
		Fn: func(ua string) chromedp.Action[chromedp.Void] {
			headers := map[string]any{"User-Agent": ua}
			return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
				if _, err := cdp.Call(ctx, t, network.Enable, network.EnableParams{}); err != nil {
					return err
				}
				_, err := cdp.Call(ctx, t, network.SetExtraHTTPHeaders, network.SetExtraHTTPHeadersParams{Headers: network.Headers(headers)})
				return err
			})
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "userAgent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36"},
		},
		Aliases: []string{"setUA", "ua", "userAgent"},
	},
	"text": {
		Desc: "Get the visible text of the first element node matching the selector (`sel`).",
		Fn: func(sel string, text *string) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.Text(sel), text)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "text", "Install the latest version of Go"},
		},
		Aliases: []string{"getText"},
	},
	"textContent": {
		Desc: "Get the text content of the first element node matching the selector (`sel`).",
		Fn: func(sel string, text *string) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.TextContent(sel), text)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "text", "Install the latest version of Go"},
		},
		Aliases: []string{"getTextContent"},
	},
	"innerHTML": {
		Desc: "Get the inner html of the first element node matching the selector (`sel`).",
		Fn: func(sel string, html *string) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.InnerHTML(sel), html)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "html", "Install the latest version of Go"},
		},
		Aliases: []string{"getInnerHTML"},
	},
	"outerHTML": {
		Desc: "Get the outer html of the first element node matching the selector (`sel`).",
		Fn: func(sel string, html *string) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.OuterHTML(sel), html)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "html", "<h1>Install the latest version of Go</h1>"},
		},
		Aliases: []string{"getOuterHTML"},
	},
	"fullHTML": {
		Desc: "Get the full html of page.",
		Fn: func(html *string) chromedp.Action[chromedp.Void] {
			expr := "new XMLSerializer().serializeToString(document);"
			return resAction(chromedp.Evaluate[string](expr), html)
		},
		Args: CDPFnArgs{
			{CDPArgTypeRes, "html", "<!DOCTYPE html><html><body><h1>hello</h1></body></html>"},
		},
		Aliases: []string{"getFullHTML", "getHTML", "html"},
	},
	"value": {
		Desc: "Get the Javascript value field of the first element node matching the selector (`sel`).",
		Fn: func(sel string, value *string) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.Value(sel), value)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "input[name=address]"},
			{CDPArgTypeRes, "value", "Fukuoka"},
		},
		Aliases: []string{"getValue"},
	},
	"setUploadFile": {
		Desc: "Set upload file (`path`) to the first element node matching the selector (`sel`).",
		Fn: func(sel, path string) chromedp.Action[chromedp.Void] {
			abs, err := filepath.Abs(path)
			if err != nil {
				return chromedp.Legacy(&errAction{err: err})
			}
			if _, err := os.Stat(abs); err != nil {
				return chromedp.Legacy(&errAction{err: err})
			}
			return chromedp.SetUploadFiles(sel, []string{abs})
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "input[name=avator]"},
			{CDPArgTypeArg, "path", "/path/to/image.png"},
		},
		Aliases: []string{"setUpload"},
	},
	"title": {
		Desc: "Get the document `title`.",
		Fn: func(title *string) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.Title(), title)
		},
		Args: CDPFnArgs{
			{CDPArgTypeRes, "title", "GitHub"},
		},
		Aliases: []string{"getTitle"},
	},
	"location": {
		Desc: "Get the document location.",
		Fn: func(url *string) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.Location(), url)
		},
		Args: CDPFnArgs{
			{CDPArgTypeRes, "url", "https://github.com"},
		},
		Aliases: []string{"getLocation"},
	},
	"attributes": {
		Desc: "Get the element attributes for the first element node matching the selector (`sel`).",
		Fn: func(sel string, attrs *map[string]string) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.Attributes(sel), attrs)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "h1"},
			{CDPArgTypeRes, "attrs", `{"class": "sr-only"}`},
		},
		Aliases: []string{"getAttributes", "attrs", "getAttrs"},
	},
	"screenshot": {
		Desc: "Take a full screenshot of the entire browser viewport.",
		Fn: func(b *[]byte) chromedp.Action[chromedp.Void] {
			return resAction(chromedp.FullScreenshot(100), b)
		},
		Args: CDPFnArgs{
			{CDPArgTypeRes, "png", "[]byte"},
		},
		Aliases: []string{"getScreenshot"},
	},
	"evaluate": {
		Desc: "Evaluate the Javascript expression (`expr`).",
		Fn: func(expr string) chromedp.Action[chromedp.Void] {
			// ignore the return value of 'evaluate'
			return chromedp.Evaluate[chromedp.Void](expr)
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "expr", `document.querySelector("h1").textContent = "hello"`},
		},
		Aliases: []string{"eval"},
	},
	"localStorage": {
		Desc: "Get localStorage items.",
		Fn: func(origin string, items *map[string]string) chromedp.Action[chromedp.Void] {
			return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
				res, err := cdp.Call(ctx, t, page.GetFrameTree, cdp.Empty{})
				if err != nil {
					return err
				}
				strageKey := domstorage.SerializedStorageKey(res.FrameTree.Frame.SecurityOrigin + "/")
				storageID := &domstorage.StorageID{
					StorageKey:     strageKey,
					IsLocalStorage: true,
				}
				resp, err := cdp.Call(ctx, t, domstorage.GetDOMStorageItems, domstorage.GetDOMStorageItemsParams{StorageID: storageID})
				if err != nil {
					return err
				}
				m := make(map[string]string)
				for _, v := range resp.Entries {
					if len(v) != 2 {
						continue
					}
					m[v[0]] = v[1]
				}
				*items = m
				return nil
			})
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "origin", "https://github.com"},
			{CDPArgTypeRes, "items", `{"key": "value"}`},
		},
		Aliases: []string{"getLocalStorage"},
	},
	"sessionStorage": {
		Desc: "Get sessionStorage items.",
		Fn: func(origin string, items *map[string]string) chromedp.Action[chromedp.Void] {
			return chromedp.Func(func(ctx context.Context, t *chromedp.Target) error {
				res, err := cdp.Call(ctx, t, page.GetFrameTree, cdp.Empty{})
				if err != nil {
					return err
				}
				strageKey := domstorage.SerializedStorageKey(res.FrameTree.Frame.SecurityOrigin + "/")
				storageID := &domstorage.StorageID{
					StorageKey:     strageKey,
					IsLocalStorage: false,
				}
				resp, err := cdp.Call(ctx, t, domstorage.GetDOMStorageItems, domstorage.GetDOMStorageItemsParams{StorageID: storageID})
				if err != nil {
					return err
				}
				m := make(map[string]string)
				for _, v := range resp.Entries {
					if len(v) != 2 {
						continue
					}
					m[v[0]] = v[1]
				}
				*items = m
				return nil
			})
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "origin", "https://github.com"},
			{CDPArgTypeRes, "items", `{"key": "value"}`},
		},
		Aliases: []string{"getSessionStorage"},
	},
}

func findCDPFn(k string) (string, CDPFn, error) {
	fn, ok := CDPFnMap[k]
	if ok {
		return k, fn, nil
	}
	for kk, fn := range CDPFnMap {
		if slices.Contains(fn.Aliases, k) {
			return kk, fn, nil
		}
	}
	return "", CDPFn{}, fmt.Errorf("not found function: %s", k)
}

func (a CDPFnArgs) ArgArgs() CDPFnArgs { //nostyle:recvtype
	res := CDPFnArgs{}
	for _, arg := range a {
		if arg.Typ == CDPArgTypeArg {
			res = append(res, arg)
		}
	}
	return res
}

func (a CDPFnArgs) ResArgs() CDPFnArgs { //nostyle:recvtype
	res := CDPFnArgs{}
	for _, arg := range a {
		if arg.Typ == CDPArgTypeRes {
			res = append(res, arg)
		}
	}
	return res
}

var (
	_ chromedp.OldAction = (*waitAction)(nil)
	_ chromedp.OldAction = (*errAction)(nil)
)

// resAction keeps the out-pointer style of CDPArgTypeRes, because evalAction allocates the result holders by reflection before the action runs.
func resAction[T any](a chromedp.Action[T], out *T) chromedp.Action[chromedp.Void] {
	return func(ctx context.Context, t *chromedp.Target) (chromedp.Void, error) {
		v, err := a(ctx, t)
		if err != nil {
			return chromedp.Void{}, err
		}
		*out = v
		return chromedp.Void{}, nil
	}
}

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

type errAction struct {
	err error
}

func (e *errAction) Do(ctx context.Context) error {
	return e.err
}
