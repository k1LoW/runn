package runn

import (
	"crypto/sha1" //#nosec G505
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/xid"
	"github.com/samber/lo"
)

// generateIDsUsingPath generates IDs using path of runbooks.
// ref: https://github.com/k1LoW/runn/blob/main/docs/designs/id.md
func generateIDsUsingPath(ops []*operator) error {
	if len(ops) == 0 {
		return nil
	}
	type tmp struct {
		o              *operator
		p              string
		rp             []string
		id             string
		normalizedPath string
	}
	var ss []*tmp
	max := 0
	root, _ := projectRoot()
	for _, o := range ops {
		p, err := filepath.Abs(filepath.Clean(o.bookPath))
		if err != nil {
			return err
		}
		rp := reversePath(p)
		ss = append(ss, &tmp{
			o:  o,
			p:  p,
			rp: rp,
		})
		if len(rp) >= max {
			max = len(rp)
		}
	}
	for i := 1; i <= max; i++ {
		var ids []string
		for _, s := range ss {
			var (
				id  string
				rp  string
				err error
			)
			if len(s.rp) < i {
				rp = strings.Join(s.rp, "/")
				id, err = generateID(rp)
				if err != nil {
					return err
				}
			} else {
				rp = strings.Join(s.rp[:i], "/")
				id, err = generateID(rp)
				if err != nil {
					return err
				}
			}
			s.id = id
			if root != "" {
				s.normalizedPath, err = filepath.Rel(root, s.p)
				if err != nil {
					return err
				}
			} else {
				s.normalizedPath = strings.Join(reversePath(rp), string(filepath.Separator))
			}
			ids = append(ids, id)
		}
		if len(lo.Uniq(ids)) == len(ss) {
			// Set ids
			for _, s := range ss {
				s.o.id = s.id
				s.o.normalizedBookPath = s.normalizedPath
			}
			return nil
		}
	}
	return errors.New("failed to generate ids")
}

func generateID(p string) (string, error) {
	if p == "" {
		return generateRandomID()
	}
	h := sha1.New() //#nosec G401
	if _, err := io.WriteString(h, p); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func generateRandomID() (string, error) {
	const prefix = "r-"
	h := sha1.New() //#nosec G401
	if _, err := io.WriteString(h, xid.New().String()); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(h.Sum(nil)), nil
}

func reversePath(p string) []string {
	return lo.Reverse(strings.Split(filepath.ToSlash(p), "/"))
}

func projectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if dir == filepath.Dir(dir) {
			return "", errors.New("failed to find project root")
		}
		if _, err := os.Stat(filepath.Join(dir, ".git", "config")); err == nil {
			return dir, nil
		}
		dir = filepath.Dir(dir)
	}
}
