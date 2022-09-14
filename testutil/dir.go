package testutil

import (
	"os"
	"path/filepath"
)

func Root() string {
	dir, _ := os.Getwd()
	for {
		if dir == "" || dir == string(filepath.Separator) {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
}

func Testdata() string {
	wd, _ := os.Getwd()
	rel, _ := filepath.Rel(wd, filepath.Join(Root(), "testdata"))
	return rel
}
