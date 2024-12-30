package runn

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/runn/testutil"
	"github.com/samber/lo"
)

func TestCDPRunner(t *testing.T) {
	if testutil.SkipCDPTest(t) {
		t.Skip("chrome not found")
	}
	hs := testutil.HTTPServer(t)
	tests := []struct {
		actions CDPActions
		wantKey string
		want    any
	}{
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]any{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "text",
					Args: map[string]any{
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
					Args: map[string]any{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn:   "fullHTML",
					Args: map[string]any{},
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
					Args: map[string]any{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "click",
					Args: map[string]any{
						"sel": "body > header > a",
					},
				},
				{
					Fn: "text",
					Args: map[string]any{
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
					Args: map[string]any{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "eval",
					Args: map[string]any{
						"expr": `document.querySelector("h1").textContent = "hello"`,
					},
				},
				{
					Fn: "text",
					Args: map[string]any{
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
					Args: map[string]any{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "attrs",
					Args: map[string]any{
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
					Args: map[string]any{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "localStorage",
					Args: map[string]any{
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
					Args: map[string]any{
						"url": fmt.Sprintf("%s/form", hs.URL),
					},
				},
				{
					Fn: "sessionStorage",
					Args: map[string]any{
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
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d/%s", i, tt.actions[0].Fn), func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)

			r, err := newCDPRunner("cc", cdpNewKey)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := r.Close(); err != nil {
					t.Error(err)
				}
			})
			t.Cleanup(func() {
				o.store.ClearSteps()
			})
			s := newStep(0, "stepKey", o, nil)
			if err := r.run(ctx, tt.actions, s); err != nil {
				t.Fatal(err)
			}
			sm := o.store.ToMap()
			sl, ok := sm["steps"].([]map[string]any)
			if !ok {
				t.Fatal("steps not found")
			}
			got, ok := sl[0][tt.wantKey]
			if !ok {
				t.Errorf("%v not found", tt.wantKey)
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestSetUploadFile(t *testing.T) {
	ctx, cancel := donegroup.WithCancel(context.Background())
	t.Cleanup(cancel)
	if testutil.SkipCDPTest(t) {
		t.Skip("chrome not found")
	}
	hs, hr := testutil.HTTPServerAndRouter(t)
	as := CDPActions{
		{
			Fn: "navigate",
			Args: map[string]any{
				"url": fmt.Sprintf("%s/form", hs.URL),
			},
		},
		{
			Fn: "setUploadFile",
			Args: map[string]any{
				"sel":  "input[name=upload0]",
				"path": "testdata/dummy.svg",
			},
		},
		{
			Fn: "click",
			Args: map[string]any{
				"sel": "input[name=submit]",
			},
		},
		{
			Fn: "wait",
			Args: map[string]any{
				"time": "1sec",
			},
		},
		{
			Fn: "text",
			Args: map[string]any{
				"sel": "h1",
			},
		},
	}
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
	s := newStep(0, "stepKey", o, nil)
	if err := r.run(ctx, as, s); err != nil {
		t.Error(err)
	}
	{
		key := "text"
		want := "Posted"
		sm := o.store.ToMap()
		sl, ok := sm["steps"].([]map[string]any)
		if !ok {
			t.Fatal("steps not found")
		}
		got, ok := sl[0][key]
		if !ok {
			t.Errorf("%v not found", key)
		}
		if diff := cmp.Diff(got, want, nil); diff != "" {
			t.Error(diff)
		}
	}
	{
		r, ok := lo.Find(hr.Requests(), func(req *http.Request) bool {
			return req.Method == http.MethodPost && req.URL.Path == "/upload"
		})
		if !ok {
			t.Fatal("not found")
		}
		f, _, err := r.FormFile("upload0")
		if err != nil {
			t.Error(err)
			return
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
			t.Error(diff)
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
			o, err := New(Book(tt.book), Var("url", ts.URL))
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
