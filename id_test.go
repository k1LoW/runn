package runn

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/hex"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samber/lo"
)

func BenchmarkReversePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
		_ = lo.Reverse(strings.Split(filepath.ToSlash(p), "/"))
	}
}

func BenchmarkSHA1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z"
		h := sha1.New()
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
		h := sha1.New()
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
