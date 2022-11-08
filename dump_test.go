package runn

import (
	"bytes"
	"context"
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
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		o.store = tt.store
		d, err := newDumpRunner(o)
		if err != nil {
			t.Fatal(err)
		}
		buf := new(bytes.Buffer)
		d.out = buf
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
	}
}
