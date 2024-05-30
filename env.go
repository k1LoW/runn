package runn

import (
	"os"

	"github.com/hashicorp/go-envparse"
)

// LoadEnvFile loads the environment variables from the given file immediately.
func LoadEnvFile(path string) error {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	parsed, err := envparse.Parse(f)
	if err != nil {
		return err
	}
	for k, v := range parsed {
		os.Setenv(k, v)
	}
	return nil
}
