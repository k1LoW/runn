package runn

import (
	"fmt"
	"os"
	"path/filepath"
)

var globalCacheDir string

// SetCacheDir set cache directory for remote runbooks
func SetCacheDir(dir string) error {
	if dir == "" {
		globalCacheDir = dir
		return nil
	}
	if globalCacheDir != "" && dir != globalCacheDir {
		return fmt.Errorf("duplicate cache dir: %s %s", dir, globalCacheDir)
	}
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%s already exists", dir)
	}

	globalCacheDir = filepath.Clean(dir)
	return nil
}

// RemoveCacheDir remove cache directory for remote runbooks
func RemoveCacheDir() error {
	if globalCacheDir == "" {
		return nil
	}
	return os.RemoveAll(globalCacheDir)
}

func cacheDir() (string, error) {
	if globalCacheDir != "" {
		if _, err := os.Stat(globalCacheDir); err != nil {
			if err := os.MkdirAll(globalCacheDir, os.ModePerm); err != nil {
				return "", err
			}
		}
		return globalCacheDir, nil
	}
	dir, err := os.MkdirTemp("", "runn")
	if err != nil {
		return "", err
	}
	globalCacheDir = dir
	return dir, nil
}
