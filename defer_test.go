package runn

import (
	"context"
	"testing"
)

func TestDeferRun(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/defer.yml"},
		{"testdata/book/defer_map.yml"},
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

			if o.useMap {
				if want := 8; len(o.store.stepMap) != want {
					t.Errorf("o.store.steps got %v, want %v", len(o.store.steps), want)
				}
			} else {
				if want := 8; len(o.store.steps) != want {
					t.Errorf("o.store.steps got %v, want %v", len(o.store.steps), want)
				}
			}
			r := o.Result()
			if want := 8; len(r.StepResults) != want {
				t.Errorf("r.StepResults got %v, want %v", len(r.StepResults), want)
			}

			t.Run("main steps", func(t *testing.T) {
				wantResults := []struct {
					desc    string
					skipped bool
					err     bool
				}{
					{"step 1", false, false},
					{"include step", false, false},
					{"step 2", false, false},
					{"step 3", false, true},
					{"step 4", true, false},
					{"defererd step c", false, false},
					{"defererd step b", false, true},
					{"defererd step a", false, false},
				}
				for i, want := range wantResults {
					got := r.StepResults[i]
					if got.Desc != want.desc {
						t.Errorf("got %v, want %v", got.Desc, want.desc)
					}
					if got.Skipped != want.skipped {
						t.Errorf("got %v, want %v", got.Skipped, want.skipped)
					}
					if (got.Err == nil) == want.err {
						t.Errorf("got %v, want %v", got.Err, want.err)
					}
				}
			})

			t.Run("include steps", func(t *testing.T) {
				wantResults := []struct {
					desc    string
					skipped bool
					err     bool
				}{
					{"included step 1", false, false},
					{"included step 2", false, false},
					{"included defererd step d", false, false},
				}

				for i, want := range wantResults {
					got := r.StepResults[1].IncludedRunResults[0].StepResults[i]
					if got.Desc != want.desc {
						t.Errorf("got %v, want %v", got.Desc, want.desc)
					}
					if got.Skipped != want.skipped {
						t.Errorf("got %v, want %v", got.Skipped, want.skipped)
					}
					if (got.Err == nil) == want.err {
						t.Errorf("got %v, want %v", got.Err, want.err)
					}
				}
			})
		})
	}
}
