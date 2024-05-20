package runn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDumpRunnerRun(t *testing.T) {
	tests := []struct {
		store     store
		expr      string
		disableNL bool
		steps     []*step
		want      string
	}{
		{
			store{},
			"'hello'",
			false,
			nil,
			`hello
`,
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			"vars.key",
			false,
			nil,
			`value
`,
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			"vars",
			false,
			nil,
			`{
  "key": "value"
}
`,
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
			false,
			nil,
			`[
  {
    "key": "value"
  }
]
`,
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
			false,
			[]*step{
				{key: "stepkey"},
				{key: "stepnext"},
			},
			`{
  "stepkey": {
    "key": "value"
  }
}
`,
		},
		{
			store{
				steps: []map[string]any{
					{"key": "value"},
				},
				vars: map[string]any{},
			},
			"steps[0]",
			false,
			nil,
			`{
  "key": "value"
}
`,
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
			false,
			[]*step{
				{key: "0"},
				{key: "1"},
			},
			`{
  "key": "value"
}
`,
		},
		{
			store{},
			"'hello'",
			true,
			nil,
			`hello`,
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
			d := newDumpRunner()
			s := newStep(0, "stepKey", o, nil)
			s.dumpRequest = &dumpRequest{
				expr:                   tt.expr,
				disableTrailingNewline: tt.disableNL,
			}
			if err := d.Run(ctx, s, true); err != nil {
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
		store     store
		expr      string
		disableNL bool
		steps     []*step
		want      string
	}{
		{
			store{},
			"'hello'",
			false,
			nil,
			`hello
`,
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			"vars.key",
			false,
			nil,
			`value
`,
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			"vars",
			false,
			nil,
			`{
  "key": "value"
}
`,
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
			false,
			nil,
			`[
  {
    "key": "value"
  }
]
`,
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
			false,
			[]*step{
				{key: "stepkey"},
				{key: "stepnext"},
			},
			`{
  "stepkey": {
    "key": "value"
  }
}
`,
		},
		{
			store{
				steps: []map[string]any{
					{"key": "value"},
				},
				vars: map[string]any{},
			},
			"steps[0]",
			false,
			nil,
			`{
  "key": "value"
}
`,
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
			false,
			[]*step{
				{key: "0"},
				{key: "1"},
			},
			`{
  "key": "value"
}
`,
		},
		{
			store{},
			"'hello'",
			true,
			nil,
			`hello`,
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
			d := newDumpRunner()
			s := newStep(0, "stepKey", o, nil)
			s.dumpRequest = &dumpRequest{
				expr:                   tt.expr,
				out:                    p,
				disableTrailingNewline: tt.disableNL,
			}
			if err := d.Run(ctx, s, true); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Errorf("got\n%#v\nwant\n%#v", string(got), tt.want)
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
			d := newDumpRunner()
			s := newStep(0, "stepKey", o, nil)
			s.dumpRequest = &dumpRequest{
				expr: "hello",
				out:  tt.out,
			}
			if err := d.Run(ctx, s, true); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(tt.want); err != nil {
				t.Fatal(err)
			}
		})
	}
}
