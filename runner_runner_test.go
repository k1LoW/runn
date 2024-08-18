package runn

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/donegroup"
	"github.com/samber/lo"
)

func TestRunnerRunner(t *testing.T) {
	type runnerKeys struct {
		http []string
		db   []string
		grpc []string
		cdp  []string
		ssh  []string
	}

	tests := []struct {
		definition map[string]any
		want       runnerKeys
		wantErr    bool
	}{
		{
			nil,
			runnerKeys{},
			true,
		},
		{
			map[string]any{
				"req": "https://example.com",
			},
			runnerKeys{
				http: []string{"req"},
				db:   []string{},
				grpc: []string{},
				cdp:  []string{},
				ssh:  []string{},
			},
			false,
		},
		{
			map[string]any{
				"req": "https://example.com",
				"db":  "sqlite3://:memory:",
			},
			runnerKeys{},
			true,
		},
		{
			map[string]any{
				"req": map[string]any{
					"endpoint": "https://example.com",
				},
			},
			runnerKeys{
				http: []string{"req"},
				db:   []string{},
				grpc: []string{},
				cdp:  []string{},
				ssh:  []string{},
			},
			false,
		},
		{
			map[string]any{
				"db": "sqlite3://:memory:",
			},
			runnerKeys{
				http: []string{},
				db:   []string{"db"},
				grpc: []string{},
				cdp:  []string{},
				ssh:  []string{},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.definition), func(t *testing.T) {
			ctx, cancel := donegroup.WithCancel(context.Background())
			t.Cleanup(cancel)
			t.Parallel()
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			rnr := newRunnerRunner()
			s := newStep(0, "stepKey", o, nil)
			s.runnerDefinition = tt.definition
			if err := rnr.Run(ctx, s); err != nil {
				if !tt.wantErr {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want error, but no error")
				return
			}
			got := runnerKeys{
				http: lo.Keys(o.httpRunners),
				db:   lo.Keys(o.dbRunners),
				grpc: lo.Keys(o.grpcRunners),
				cdp:  lo.Keys(o.cdpRunners),
				ssh:  lo.Keys(o.sshRunners),
			}
			opts := []cmp.Option{
				cmp.AllowUnexported(runnerKeys{}),
			}
			if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}
