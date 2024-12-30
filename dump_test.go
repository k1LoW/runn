package runn

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/maskedio"
	"github.com/k1LoW/runn/internal/store"
)

func TestDumpRunnerRun(t *testing.T) {
	tests := []struct {
		store     *store.Store
		expr      string
		disableNL bool
		steps     []*step
		want      string
	}{
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				return s
			}(),
			"'hello'",
			false,
			nil,
			`hello
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			"vars.key",
			false,
			nil,
			`value
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			"vars",
			false,
			nil,
			`{
  "key": "value"
}
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, true, []string{"stepkey", "stepnext"})
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
			"steps[0]",
			false,
			nil,
			`{
  "key": "value"
}
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, true, []string{"0", "1"})
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				return s
			}(),
			"'hello'",
			true,
			nil,
			`hello`,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d.%s", i, tt.expr), func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			buf := new(bytes.Buffer)
			o.store = tt.store
			o.stdout = maskedio.NewWriter(buf)
			sm := tt.store.ToMap()
			_, useMap := sm["steps"].(map[string]any)
			o.useMap = useMap
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
		store     *store.Store
		expr      string
		disableNL bool
		steps     []*step
		want      string
	}{
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				return s
			}(),
			"'hello'",
			false,
			nil,
			`hello
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			"vars.key",
			false,
			nil,
			`value
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			"vars",
			false,
			nil,
			`{
  "key": "value"
}
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, true, []string{"stepkey", "stepnext"})
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
			"steps[0]",
			false,
			nil,
			`{
  "key": "value"
}
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, true, []string{"0", "1"})
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
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
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				return s
			}(),
			"'hello'",
			true,
			nil,
			`hello`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d.%s with out", i, tt.expr), func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			p := filepath.Join(t.TempDir(), "tmp")
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			o.store = tt.store
			sm := tt.store.ToMap()
			_, useMap := sm["steps"].(map[string]any)
			o.useMap = useMap
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
		store *store.Store
		out   string
		want  string
	}{
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				return s
			}(),
			rp,
			fp,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				return s
			}(),
			filepath.Join(tmp, "temp2"),
			filepath.Join(tmp, "temp2"),
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", filepath.Join(tmp, "value"))
				return s
			}(),
			"{{ vars.key }}",
			filepath.Join(tmp, "value"),
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", filepath.Join(tmp, "value2"))
				return s
			}(),
			"{{ vars.key + '.ext' }}",
			filepath.Join(tmp, "value2.ext"),
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", filepath.Join(tmp, "value3"))
				return s
			}(),
			"{{ vars.key }}.ext",
			filepath.Join(tmp, "value3.ext"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.out, func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
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

func TestDumpRunnerRunWithSecrets(t *testing.T) {
	tests := []struct {
		store                 *store.Store
		expr                  string
		steps                 []*step
		secrets               []string
		disableMaskingSecrets bool
		want                  string
	}{
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			"vars.key",
			nil,
			[]string{"vars.key"},
			false,
			`*****
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			"vars",
			nil,
			[]string{"vars.key"},
			false,
			`{
  "key": "*****"
}
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
			"steps",
			nil,
			[]string{"steps[0].key"},
			false,
			`[
  {
    "key": "*****"
  }
]
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, true, []string{"stepkey", "stepnext"})
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
			"steps",
			[]*step{
				{key: "stepkey"},
				{key: "stepnext"},
			},
			[]string{"steps.stepkey.key"},
			false,
			`{
  "stepkey": {
    "key": "*****"
  }
}
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, false, nil)
				s.SetRunNIndex(0)
				s.SetVar("key", "value")
				return s
			}(),
			"vars.key",
			nil,
			[]string{"vars.key"},
			true,
			`value
`,
		},
		{
			func() *store.Store {
				s := store.New(map[string]any{}, map[string]any{}, nil, true, []string{"stepkey", "stepnext"})
				s.SetRunNIndex(0)
				s.Record(0, map[string]any{"key": "value"})
				return s
			}(),
			"steps",
			[]*step{
				{key: "stepkey"},
				{key: "stepnext"},
			},
			[]string{"current.key"},
			false,
			`{
  "stepkey": {
    "key": "*****"
  }
}
`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d.%s", i, tt.expr), func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			buf := new(bytes.Buffer)
			o, err := New(Stdout(buf))
			if err != nil {
				t.Fatal(err)
			}
			tt.store.SetSecrets(tt.secrets)
			tt.store.SetMaskRule(o.store.MaskRule())
			tt.store.SetMaskKeywords(tt.store.ToMap())
			o.store = tt.store
			sm := tt.store.ToMap()
			_, useMap := sm["steps"].(map[string]any)
			o.useMap = useMap
			o.steps = tt.steps
			d := newDumpRunner()
			s := newStep(0, "stepKey", o, nil)
			s.dumpRequest = &dumpRequest{
				expr:                  tt.expr,
				disableMaskingSecrets: tt.disableMaskingSecrets,
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
