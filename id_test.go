package runn

import (
	"crypto/sha1" //#nosec G505
	"encoding/base32"
	"encoding/hex"
	"io"
	"path/filepath"
	"strings"
	"testing"

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

func BenchmarkReversePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
		_ = lo.Reverse(strings.Split(filepath.ToSlash(p), "/"))
	}
}

func BenchmarkSHA1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
		h := sha1.New() //#nosec G401
		_, _ = io.WriteString(h, p)
		_ = hex.EncodeToString(h.Sum(nil))
	}
}

func BenchmarkBase32(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
		_ = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(p))
	}
}

func BenchmarkReverseAndHashBySHA1Path(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
		r := lo.Reverse(strings.Split(filepath.ToSlash(p), "/"))
		h := sha1.New() //#nosec G401
		_, _ = io.WriteString(h, strings.Join(r[0:5], "/"))
		_ = hex.EncodeToString(h.Sum(nil))
	}
}

func BenchmarkReverseAndEncodeByBase32Path(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
		r := lo.Reverse(strings.Split(filepath.ToSlash(p), "/"))
		_ = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString([]byte(strings.Join(r[0:5], "/")))
	}
}
