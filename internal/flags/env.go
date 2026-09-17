package flags

import (
	"os"

	"github.com/hashicorp/go-envparse"
)

// loadEnvFile loads the environment variables from the given file into the process environment.
// It is a startup step of the CLI, as if the variables had been exported in the shell before invoking runn, so it must run before any operator is created.
func loadEnvFile(path string) error {
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
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}
