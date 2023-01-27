//go:build loadt

package runn

import (
	"context"
	"testing"
	"time"

	"github.com/ryo-yamaoka/otchkiss"
	"github.com/ryo-yamaoka/otchkiss/setting"
)

func TestLoadt(t *testing.T) {
	tests := []struct {
		in        string
		concarent int
	}{
		{"testdata/book/include_main.yml", 2},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			opts := []Option{}
			o, err := Load(tt.in, opts...)
			if err != nil {
				t.Error(err)
			}
			s, err := setting.New(tt.concarent, 5*time.Second, 5*time.Second)
			if err != nil {
				t.Error(err)
			}
			ot, err := otchkiss.FromConfig(o, s, 100_000_000)
			if err != nil {
				t.Error(err)
			}
			if err := ot.Start(context.Background()); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCheckThreshold(t *testing.T) {
	tests := []struct {
		lr        *loadtResult
		threshold string
		wantErr   bool
	}{
		{&loadtResult{}, "", false},
		{&loadtResult{succeeded: 11}, "succeeded > 10", false},
		{&loadtResult{failed: 10}, "failed < 10", true},
	}
	for _, tt := range tests {
		t.Run(tt.threshold, func(t *testing.T) {
			err := tt.lr.CheckThreshold(tt.threshold)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("got err: %s", err)
			}
			if tt.wantErr {
				t.Error("want error")
			}
		})
	}
}
