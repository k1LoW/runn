package runn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDumpRunnerRun(t *testing.T) {
	tests := []struct {
		store store
		expr  string
		want  string
	}{
		{
			store{},
			"'hello'",
			`hello
`,
		},
		{
			store{
				steps: []map[string]interface{}{},
				vars: map[string]interface{}{
					"key": "value",
				},
			},
			"vars.key",
			`value
`,
		},
		{
			store{
				steps: []map[string]interface{}{},
				vars: map[string]interface{}{
					"key": "value",
				},
			},
			"vars",
			`{
  "key": "value"
}
`,
		},
		{
			store{
				steps: []map[string]interface{}{
					map[string]interface{}{
						"key": "value",
					},
				},
				vars: map[string]interface{}{},
			},
			"steps",
			`[
  {
    "key": "value"
  }
]
`,
		},
		{
			store{
				stepMap: map[string]map[string]interface{}{
					"stepkey": {"key": "value"},
				},
				vars:   map[string]interface{}{},
				useMap: true,
			},
			"steps",
			`{
  "stepkey": {
    "key": "value"
  }
}
`,
		},
		{
			store{
				steps: []map[string]interface{}{
					{"key": "value"},
				},
				vars: map[string]interface{}{},
			},
			"steps[0]",
			`{
  "key": "value"
}
`,
		},
		{
			store{
				stepMap: map[string]map[string]interface{}{
					"0": {"key": "value"},
				},
				vars:   map[string]interface{}{},
				useMap: true,
			},
			"steps['0']",
			`{
  "key": "value"
}
`,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			buf := new(bytes.Buffer)
			o.store = tt.store
			o.stdout = buf
			d, err := newDumpRunner(o)
			if err != nil {
				t.Fatal(err)
			}
			req := &dumpRequest{
				expr: tt.expr,
			}
			if err := d.Run(ctx, req); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			if got != tt.want {
				t.Errorf("got\n%#v\nwant\n%#v", got, tt.want)
			}
		})

		t.Run(fmt.Sprintf("%s with out", tt.expr), func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "tmp")
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			o.store = tt.store
			d, err := newDumpRunner(o)
			if err != nil {
				t.Fatal(err)
			}
			req := &dumpRequest{
				expr: tt.expr,
				out:  p,
			}
			if err := d.Run(ctx, req); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != strings.TrimSuffix(tt.want, "\n") {
				t.Errorf("got\n%#v\nwant\n%#v", string(got), strings.TrimSuffix(tt.want, "\n"))
			}
		})
	}
}
