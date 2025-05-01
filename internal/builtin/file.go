package builtin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func File(root string) func(string) (any, error) {
	return func(path string) (any, error) {
		var scheme string
		if strings.Contains(path, "://") {
			splitted := strings.SplitN(path, "://", 2)
			if len(splitted) != 2 {
				return nil, fmt.Errorf("invalid file path: %s", path)
			}
			scheme = splitted[0]
			path = splitted[1]
		} else {
			scheme = "file"
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(root, path)
		}
		if fi, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, err
		} else {
			if fi.IsDir() {
				return nil, fmt.Errorf("path is a directory: %s", path)
			}
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		switch scheme {
		case "binary":
			return b, nil
		case "file":
			return string(b), nil
		case "json":
			var v any
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		default:
			return nil, fmt.Errorf("unsupported file scheme: %s", scheme)
		}
	}
}
