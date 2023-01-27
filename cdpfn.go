package runn

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	"latestTab": {
		Desc: "Change current frame to latest tab.",
		Fn: func() chromedp.Action {
			// dummy
			return nil
		},
		Args:    CDPFnArgs{},
		Aliases: []string{"latestTarget"},
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
	"scroll": {
		Desc: "Scroll the window to the first element node matching the selector (`sel`).",
		Fn:   chromedp.ScrollIntoView,
		Args: CDPFnArgs{
			{CDPArgTypeArg, "sel", "body > footer"},
		},
		Aliases: []string{"scrollIntoView"},
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
	"setUserAgent": {
		Desc: "Set the default User-Agent",
		Fn: func(ua string) []chromedp.Action {
			headers := map[string]interface{}{"User-Agent": ua}
			return []chromedp.Action{
				network.Enable(),
				network.SetExtraHTTPHeaders(network.Headers(headers)),
			}
		},
		Args: CDPFnArgs{
			{CDPArgTypeArg, "userAgent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36"},
		},
		Aliases: []string{"setUA", "ua", "userAgent"},
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
	"fullHTML": {
		Desc: "Get the full html of page.",
		Fn: func(html *string) chromedp.Action {
			expr := "new XMLSerializer().serializeToString(document);"
			return chromedp.Evaluate(expr, html)
		},
		Args: CDPFnArgs{
			{CDPArgTypeRes, "html", "<!DOCTYPE html><html><body><h1>hello</h1></body></html>"},
		},
		Aliases: []string{"getFullHTML", "getHTML", "html"},
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
	"setUploadFile": {
		Desc: "Set upload file (`path`) to the first element node matching the selector (`sel`).",
		Fn: func(sel, path string) chromedp.Action {
			abs, err := filepath.Abs(path)
			if err != nil {
				return &errAction{err: err}
			}
			if err := fetchFile(abs); err != nil {
				return &errAction{err: err}
			}
			if _, err := os.Stat(abs); err != nil {
				return &errAction{err: err}
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
	"screenshot": {
		Desc: "Take a full screenshot of the entire browser viewport.",
		Fn: func(b *[]byte) chromedp.Action {
			return chromedp.FullScreenshot(b, 100)
		},
		Args: CDPFnArgs{
			{CDPArgTypeRes, "png", "[]byte"},
		},
		Aliases: []string{"getScreenshot"},
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
	"localStorage": {
		Desc: "Get localStorage items.",
		Fn: func(origin string, items *map[string]string) chromedp.Action {
			return chromedp.ActionFunc(func(ctx context.Context) error {
				frameTree, err := page.GetFrameTree().Do(ctx)
				if err != nil {
					return err
				}
				strageKey := domstorage.SerializedStorageKey(frameTree.Frame.SecurityOrigin + "/")
				storageID := &domstorage.StorageID{
					StorageKey:     strageKey,
					IsLocalStorage: true,
				}
				resp, err := domstorage.GetDOMStorageItems(storageID).Do(ctx)
				if err != nil {
					return err
				}
				m := make(map[string]string)
				for _, v := range resp {
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
		Fn: func(origin string, items *map[string]string) chromedp.Action {
			return chromedp.ActionFunc(func(ctx context.Context) error {
				frameTree, err := page.GetFrameTree().Do(ctx)
				if err != nil {
					return err
				}
				strageKey := domstorage.SerializedStorageKey(frameTree.Frame.SecurityOrigin + "/")
				storageID := &domstorage.StorageID{
					StorageKey:     strageKey,
					IsLocalStorage: false,
				}
				resp, err := domstorage.GetDOMStorageItems(storageID).Do(ctx)
				if err != nil {
					return err
				}
				m := make(map[string]string)
				for _, v := range resp {
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
		for _, a := range fn.Aliases {
			if a == k {
				return kk, fn, nil
			}
		}
	}
	return "", CDPFn{}, fmt.Errorf("not found function: %s", k)
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

var (
	_ chromedp.Action = (*waitAction)(nil)
	_ chromedp.Action = (*errAction)(nil)
)

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
