package runn

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/samber/lo"
)

func TestGenerateIDsUsingPath(t *testing.T) {
	tests := []struct {
		paths            []string
		seedReversePaths []string
	}{
		{
			[]string{"a.yml", "b.yml", "c.yml"},
			[]string{"a.yml", "b.yml", "c.yml"},
		},
		{
			[]string{"path/to/a.yml", "path/to/b.yml", "path/to/c.yml"},
			[]string{"a.yml", "b.yml", "c.yml"},
		},
		{
			[]string{"path/to/bb/a.yml", "path/to/aa/a.yml"},
			[]string{"a.yml/bb", "a.yml/aa"},
		},
		{
			[]string{"path/to/bb/a.yml", "../../path/to/aa/a.yml"},
			[]string{"a.yml/bb", "a.yml/aa"},
		},
	}
	for _, tt := range tests {
		om := map[string]*operator{}
		for _, p := range tt.paths {
			om[p] = &operator{
				bookPath: p,
			}
		}
		ops := &operatorN{
			om: om,
		}
		if err := ops.generateIDsUsingPath(); err != nil {
			t.Fatal(err)
		}
		got := lo.Map(lo.Values(ops.om), func(item *operator, _ int) string {
			return item.id
		})
		want := lo.Map(tt.seedReversePaths, func(item string, _ int) string {
			id, err := generateID(item)
			if err != nil {
				t.Fatal(err)
			}
			return id
		})
		slices.Sort(got)
		slices.Sort(want)

		if diff := cmp.Diff(want, got); diff != "" {
			t.Error(diff)
		}
	}
}
