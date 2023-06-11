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
		steps []*step
	}{
		{
			store{},
			"'hello'",
			`hello
`,
			nil,
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			"vars.key",
			`value
`,
			nil,
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			"vars",
			`{
  "key": "value"
}
`,
			nil,
		},
		{
			store{
				steps: []map[string]any{
					{
						"key": "value",
					},
				},
				vars: map[string]any{},
			},
			"steps",
			`[
  {
    "key": "value"
  }
]
`,
			nil,
		},
		{
			store{
				steps: []map[string]any{},
				stepMap: map[string]map[string]any{
					"stepkey": {"key": "value"},
				},
				vars:   map[string]any{},
				useMap: true,
			},
			"steps",
			`{
  "stepkey": {
    "key": "value"
  }
}
`,
			[]*step{
				{key: "stepkey"},
				{key: "stepnext"},
			},
		},
		{
			store{
				steps: []map[string]any{
					{"key": "value"},
				},
				vars: map[string]any{},
			},
			"steps[0]",
			`{
  "key": "value"
}
`,
			nil,
		},
		{
			store{
				stepMap: map[string]map[string]any{
					"0": {"key": "value"},
				},
				vars:   map[string]any{},
				useMap: true,
			},
			"steps['0']",
			`{
  "key": "value"
}
`,
			[]*step{
				{key: "0"},
				{key: "1"},
			},
		},
	}
	ctx := context.Background()
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d.%s", i, tt.expr), func(t *testing.T) {
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			buf := new(bytes.Buffer)
			o.store = tt.store
			o.stdout = buf
			o.useMap = tt.store.useMap
			o.steps = tt.steps
			d, err := newDumpRunner(o)
			if err != nil {
				t.Fatal(err)
			}
			req := &dumpRequest{
				expr: tt.expr,
			}
			if err := d.Run(ctx, req, true); err != nil {
				t.Fatal(err)
			}
			got := buf.String()
			if got != tt.want {
				t.Errorf("got\n%#v\nwant\n%#v", got, tt.want)
			}
		})
	}
}

func TestDumpRunnerRunWithOut(t *testing.T) {
	tests := []struct {
		store store
		expr  string
		want  string
		steps []*step
	}{
		{
			store{},
			"'hello'",
			`hello
`,
			nil,
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			"vars.key",
			`value
`,
			nil,
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			"vars",
			`{
  "key": "value"
}
`,
			nil,
		},
		{
			store{
				steps: []map[string]any{
					{
						"key": "value",
					},
				},
				vars: map[string]any{},
			},
			"steps",
			`[
  {
    "key": "value"
  }
]
`,
			nil,
		},
		{
			store{
				steps: []map[string]any{},
				stepMap: map[string]map[string]any{
					"stepkey": {"key": "value"},
				},
				vars:   map[string]any{},
				useMap: true,
			},
			"steps",
			`{
  "stepkey": {
    "key": "value"
  }
}
`,
			[]*step{
				{key: "stepkey"},
				{key: "stepnext"},
			},
		},
		{
			store{
				steps: []map[string]any{
					{"key": "value"},
				},
				vars: map[string]any{},
			},
			"steps[0]",
			`{
  "key": "value"
}
`,
			nil,
		},
		{
			store{
				stepMap: map[string]map[string]any{
					"0": {"key": "value"},
				},
				vars:   map[string]any{},
				useMap: true,
			},
			"steps['0']",
			`{
  "key": "value"
}
`,
			[]*step{
				{key: "0"},
				{key: "1"},
			},
		},
	}
	ctx := context.Background()
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d.%s with out", i, tt.expr), func(t *testing.T) {
			p := filepath.Join(t.TempDir(), "tmp")
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			o.store = tt.store
			o.useMap = tt.store.useMap
			o.steps = tt.steps
			d, err := newDumpRunner(o)
			if err != nil {
				t.Fatal(err)
			}
			req := &dumpRequest{
				expr: tt.expr,
				out:  p,
			}
			if err := d.Run(ctx, req, true); err != nil {
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

func TestDumpRunnerRunWithExpandOut(t *testing.T) {
	tmp := t.TempDir()
	fp := filepath.Join(tmp, "tmp")
	cd, err := filepath.Abs(".")
	if err != nil {
		t.Error(err)
	}
	rp, err := filepath.Rel(cd, fp)
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		store store
		out   string
		want  string
	}{
		{
			store{},
			rp,
			fp,
		},
		{
			store{},
			filepath.Join(tmp, "temp2"),
			filepath.Join(tmp, "temp2"),
		},
		{
			store{
				vars: map[string]any{
					"key": filepath.Join(tmp, "value"),
				},
			},
			"{{ vars.key }}",
			filepath.Join(tmp, "value"),
		},
		{
			store{
				vars: map[string]any{
					"key": filepath.Join(tmp, "value2"),
				},
			},
			"{{ vars.key + '.ext' }}",
			filepath.Join(tmp, "value2.ext"),
		},
		{
			store{
				vars: map[string]any{
					"key": filepath.Join(tmp, "value3"),
				},
			},
			"{{ vars.key }}.ext",
			filepath.Join(tmp, "value3.ext"),
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.out, func(t *testing.T) {
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
				expr: "hello",
				out:  tt.out,
			}
			if err := d.Run(ctx, req, true); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(tt.want); err != nil {
				t.Fatal(err)
			}
		})
	}
}
