package runn

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

func Paths(pathp string) ([]string, error) {
	paths := []string{}
	listp := filepath.SplitList(pathp)
	for _, pp := range listp {
		base, pattern := doublestar.SplitPattern(pp)
		abs, err := filepath.Abs(base)
		if err != nil {
			return nil, err
		}
		fsys := os.DirFS(abs)
		if err := doublestar.GlobWalk(fsys, pattern, func(p string, d fs.DirEntry) error {
			if d.IsDir() {
				return nil
			}
			paths = append(paths, filepath.Join(base, p))
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return unique(paths), nil
}

func unique(in []string) []string {
	u := []string{}
	m := map[string]struct{}{}
	for _, s := range in {
		if _, ok := m[s]; ok {
			continue
		}
		u = append(u, s)
		m[s] = struct{}{}
	}
	return u
}
