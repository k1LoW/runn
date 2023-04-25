package runn

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/runn/testutil"
)

func TestCDPRunner(t *testing.T) {
	if testutil.SkipCDPTest(t) {
		t.Skip("chrome not found")
	}
	hs := testutil.HTTPServer(t)
	tests := []struct {
		actions CDPActions
		wantKey string
		want    interface{}
	}{
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "text",
					Args: map[string]interface{}{
						"sel": "h1",
					},
				},
			},
			"text",
			"Test Form",
		},
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn:   "fullHTML",
					Args: map[string]interface{}{},
				},
			},
			"html",
			`<!DOCTYPE html><html xmlns="http://www.w3.org/1999/xhtml"><head>
  <title>For runn test</title>
</head>
<body>
  <header>
    <h1 class="runn-test" data-test-id="runn-h1">Test Form</h1>
    <a href="/hello">Link</a>
  </header>
  <form class="form-test" method="POST" action="/upload" enctype="multipart/form-data">
    <input name="username" type="text" />
    <input name="upload0" type="file" />
    <input name="upload1" type="file" />
    <input name="submit" type="submit" />
  </form>

  <input id="newtab" type="button" value="open" onclick="window.open(&quot;/hello&quot;, &quot;_blank&quot;);" />

  <script>
	  localStorage.setItem('local', 'storage');
	  sessionStorage.setItem('session', 'storage');
  </script>


</body></html>`,
		},
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "click",
					Args: map[string]interface{}{
						"sel": "body > header > a",
					},
				},
				{
					Fn: "text",
					Args: map[string]interface{}{
						"sel": "h1",
					},
				},
			},
			"text",
			"Hello",
		},
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "eval",
					Args: map[string]interface{}{
						"expr": `document.querySelector("h1").textContent = "hello"`,
					},
				},
				{
					Fn: "text",
					Args: map[string]interface{}{
						"sel": "h1",
					},
				},
			},
			"text",
			"hello",
		},
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "attrs",
					Args: map[string]interface{}{
						"sel": "h1",
					},
				},
			},
			"attrs",
			map[string]string{
				"class":        "runn-test",
				"data-test-id": "runn-h1",
			},
		},
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "localStorage",
					Args: map[string]interface{}{
						"origin": hs.URL,
					},
				},
			},
			"items",
			map[string]string{
				"local": "storage",
			},
		},
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "sessionStorage",
					Args: map[string]interface{}{
						"origin": hs.URL,
					},
				},
			},
			"items",
			map[string]string{
				"session": "storage",
			},
		},
	}
	ctx := context.Background()
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d/%s", i, tt.actions[0].Fn), func(t *testing.T) {
			r, err := newCDPRunner("cc", cdpNewKey)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := r.Close(); err != nil {
					t.Error(err)
				}
			})
			r.operator = o
			t.Cleanup(func() {
				o.store.steps = []map[string]interface{}{}
			})
			if err := r.Run(ctx, tt.actions); err != nil {
				t.Fatal(err)
			}
			got, ok := o.store.steps[0][tt.wantKey]
			if !ok {
				t.Errorf("%v not found", tt.wantKey)
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func TestSetUploadFile(t *testing.T) {
	if testutil.SkipCDPTest(t) {
		t.Skip("chrome not found")
	}
	hs, hr := testutil.HTTPServerAndRouter(t)
	as := CDPActions{
		{
			Fn: "navigate",
			Args: map[string]interface{}{
				"url": fmt.Sprintf("%s/form", hs.URL),
			},
		},
		{
			Fn: "setUploadFile",
			Args: map[string]interface{}{
				"sel":  "input[name=upload0]",
				"path": "testdata/dummy.svg",
			},
		},
		{
			Fn: "click",
			Args: map[string]interface{}{
				"sel": "input[name=submit]",
			},
		},
		{
			Fn: "text",
			Args: map[string]interface{}{
				"sel": "h1",
			},
		},
	}
	ctx := context.Background()
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	r, err := newCDPRunner("cc", cdpNewKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	})
	r.operator = o
	if err := r.Run(ctx, as); err != nil {
		t.Error(err)
	}
	{
		key := "text"
		want := "Posted"
		got, ok := o.store.steps[0][key]
		if !ok {
			t.Errorf("%v not found", key)
		}
		if diff := cmp.Diff(got, want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
	{
		r := hr.Requests()[1]
		f, _, err := r.FormFile("upload0")
		if err != nil {
			t.Error(err)
		}
		t.Cleanup(func() {
			_ = f.Close()
		})
		got, err := io.ReadAll(f)
		if err != nil {
			t.Error(err)
		}
		want, err := os.ReadFile("testdata/dummy.svg")
		if err != nil {
			t.Error(err)
		}
		if diff := cmp.Diff(got, want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestCDP(t *testing.T) {
	if testutil.SkipCDPTest(t) {
		t.Skip("chrome not found")
	}
	tests := []struct {
		book string
	}{
		{"testdata/book/cdp.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.book, func(t *testing.T) {
			t.Parallel()
			ts := testutil.HTTPServer(t)
			o, err := New(Book(tt.book), Var("url", fmt.Sprintf("%s", ts.URL)))
			if err != nil {
				t.Fatal(err)
			}
			for _, r := range o.cdpRunners {
				// override timeoutByStep
				r.timeoutByStep = 2 * time.Second
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}
