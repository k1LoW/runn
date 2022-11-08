package runn

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

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
	}
	ctx := context.Background()
	o, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
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
				t.Error(err)
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
				"sel":  "input[name=upload]",
				"path": "docs/logo.svg",
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
		f, _, err := r.FormFile("upload")
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
		want, err := os.ReadFile("docs/logo.svg")
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
			hs := testutil.HTTPServer(t)
			o, err := New(Book(tt.book), Var("url", fmt.Sprintf("%s/form", hs.URL)))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}
