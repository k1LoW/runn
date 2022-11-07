package runn

import (
	"context"
	"fmt"
	"testing"

	"github.com/k1LoW/runn/testutil"
)

func TestCDPRunner(t *testing.T) {
	if testutil.SkipCDPTest(t) {
		t.Skip("chrome not found")
	}
	tests := []struct {
		actions CDPActions
		want    string
	}{
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": "https://pkg.go.dev/time",
					},
				},
				{
					Fn: "text",
					Args: map[string]interface{}{
						"sel": "h1",
					},
				},
			},
			"time",
		},
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": "https://pkg.go.dev/time",
					},
				},
				{
					Fn: "click",
					Args: map[string]interface{}{
						"sel": "body > header > div.go-Header-inner > nav > div > ul > li:nth-child(2) > a",
					},
				},
				{
					Fn: "waitVisible",
					Args: map[string]interface{}{
						"sel": "body > footer",
					},
				},
				{
					Fn: "text",
					Args: map[string]interface{}{
						"sel": "h1",
					},
				},
			},
			"Install the latest version of Go",
		},
		{
			CDPActions{
				{
					Fn: "navigate",
					Args: map[string]interface{}{
						"url": "https://pkg.go.dev/time",
					},
				},
				{
					Fn: "evaluate",
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
			"hello",
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
			got := o.store.steps[0]["text"]
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
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
			o, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}
