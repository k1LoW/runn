package runn

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	t.Setenv("TEST_LOAD_ENV", "")
	tests := []struct {
		envs    string
		wantEnv string
	}{
		{"", ""},
		{"TEST_LOAD_ENV=hoge", "hoge"},
		{"TEST_LOAD_ENV=hoge\n", "hoge"},
		{"TEST_LOAD_ENV=hoge\nTEST_LOAD_ENV=fuga", "fuga"},
	}
	for _, tt := range tests {
		t.Run(tt.envs, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), ".env")
			if err := os.WriteFile(path, []byte(tt.envs), 0600); err != nil {
				t.Fatal(err)
			}
			if err := LoadEnvFile(path); err != nil {
				t.Fatal(err)
			}
			if got := os.Getenv("TEST_LOAD_ENV"); got != tt.wantEnv {
				t.Errorf("got %v\nwant %v", got, tt.wantEnv)
			}
		})
	}
}
