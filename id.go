package runn

import (
	"crypto/sha1" //#nosec G505
	"encoding/hex"
	"errors"
	"io"
	"path/filepath"
	"strings"

	"github.com/rs/xid"
	"github.com/samber/lo"
)

// generateIDsUsingPath generates IDs using path of runbooks.
// ref: https://github.com/k1LoW/runn/blob/main/docs/designs/id.md
func (opn *operatorN) generateIDsUsingPath() error {
	// Use ops.om. because om caches all the operatorN it has traversed so far.
	if len(opn.om) == 0 {
		return nil
	}
	ops := lo.Values(opn.om)
	type tmp struct {
		o  *operator
		p  string
		rp []string
		id string
	}
	var ss []*tmp
	max := 0
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
				err error
			)
			if len(s.rp) < i {
				id, err = generateID(strings.Join(s.rp, "/"))
				if err != nil {
					return err
				}
			} else {
				id, err = generateID(strings.Join(s.rp[:i], "/"))
				if err != nil {
					return err
				}
			}
			s.id = id
			ids = append(ids, id)
		}
		if len(lo.Uniq(ids)) == len(ss) {
			// Set ids
			for _, s := range ss {
				s.o.id = s.id
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
