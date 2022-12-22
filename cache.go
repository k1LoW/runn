package runn

import (
	"fmt"
	"os"
)

var globalCacheDir string

func SetCacheDir(dir string) error {
	if globalCacheDir != "" && dir != globalCacheDir {
		return fmt.Errorf("duplicate cache dir: %s %s", dir, globalCacheDir)
	}
	globalCacheDir = dir
	return nil
}

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
