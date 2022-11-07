//go:build integration

package runn

import (
	"context"
	"fmt"
	"testing"
)

func TestCDPRunner(t *testing.T) {
	tests := []struct {
		actions cdpActions
		want    string
	}{
		{
			cdpActions{
				{
					fn: "navigate",
					args: map[string]interface{}{
						"url": "https://pkg.go.dev/time",
					},
				},
				{
					fn: "text",
					args: map[string]interface{}{
						"sel": "h1",
					},
				},
			},
			"time",
		},
		{
			cdpActions{
				{
					fn: "navigate",
					args: map[string]interface{}{
						"url": "https://pkg.go.dev/time",
					},
				},
				{
					fn: "click",
					args: map[string]interface{}{
						"sel": "body > header > div.go-Header-inner > nav > div > ul > li:nth-child(2) > a",
					},
				},
				{
					fn: "waitVisible",
					args: map[string]interface{}{
						"sel": "body > footer",
					},
				},
				{
					fn: "text",
					args: map[string]interface{}{
						"sel": "h1",
					},
				},
			},
			"Install the latest version of Go",
		},
		{
			cdpActions{
				{
					fn: "navigate",
					args: map[string]interface{}{
						"url": "https://pkg.go.dev/time",
					},
				},
				{
					fn: "evaluate",
					args: map[string]interface{}{
						"expr": `document.querySelector("h1").textContent = "hello"`,
					},
				},
				{
					fn: "text",
					args: map[string]interface{}{
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
