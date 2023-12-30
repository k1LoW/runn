package runn

import (
	"testing"
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
		var ops []*operator
		for _, p := range tt.paths {
			ops = append(ops, &operator{
				bookPath: p,
			})
		}
		if err := generateIDsUsingPath(ops); err != nil {
			t.Fatal(err)
		}
		for i, o := range ops {
			want, err := generateID(tt.seedReversePaths[i])
			if err != nil {
				t.Fatal(err)
			}
			if o.id != want {
				t.Errorf("want %s, got %s", want, o.id)
			}
		}
	}
}
