package builtin

import (
	"errors"
	"os"

	"github.com/k1LoW/runn/internal/fs"
)

// File returns a function that reads a file from the given root directory.
func File(root string) func(string) (any, error) {
	return func(path string) (any, error) {
		p, err := fs.Path(path, root)
		if err != nil {
			return nil, err
		}
		b, err := fs.ReadFile(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, nil
			}
			return nil, err
		}
		return string(b), nil
	}
}
