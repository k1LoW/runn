package runn

import (
	"context"
	"testing"
)

func TestForceRun(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/force.yml"},
		{"testdata/book/force_step.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			ctx := context.Background()
			o, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err == nil {
				t.Fatal("expected error")
			}
			for _, sr := range o.Result().StepResults {
				if sr.Skipped {
					t.Errorf("got %v, want %v", sr.Skipped, false)
				}
			}
		})
	}
}

func TestIncludeForceRun(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/force_include.yml"},
		{"testdata/book/inline_include_with_force.yml"},
	}
	for _, tt := range tests {
		t.Run(tt.book, func(t *testing.T) {
			ctx := context.Background()
			o, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err == nil {
				t.Fatal("expected error")
			}
			r := o.Result()
			if len(r.StepResults) == 0 {
				t.Fatal("expected step results")
			}
			foundIncludedRunResults := false
			for _, sr := range r.StepResults {
				if len(sr.IncludedRunResults) == 0 {
					continue
				}
				foundIncludedRunResults = true
				for j, irr := range sr.IncludedRunResults {
					if len(irr.StepResults) == 0 {
						t.Fatalf("included run[%d]: expected step results", j)
					}
					for i, isr := range irr.StepResults {
						if isr.Skipped {
							t.Errorf("included step[%d]: got skipped=%v, want skipped=false", i, isr.Skipped)
						}
					}
				}
			}
			if !foundIncludedRunResults {
				t.Fatal("expected included run results")
			}
		})
	}
}
