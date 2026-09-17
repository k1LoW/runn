package runn

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/runn/internal/scope"
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

func TestEnvSnapshotTakenAtRunStart(t *testing.T) {
	const key = "TEST_RUN_ENV_SNAPSHOT"
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "book.yml")
	if err := os.WriteFile(path, []byte(`
desc: Tests for environment variable snapshot.
steps:
  -
    test: env.`+key+` == vars.want
`), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(key, "before-new")
	o, err := New(Book(path), Var("want", "at-run-start"), Scopes(scope.AllowReadParent))
	if err != nil {
		t.Fatal(err)
	}
	// The environment set after the operator is created is still reflected.
	t.Setenv(key, "at-run-start")
	if err := o.Run(ctx); err != nil {
		t.Error(err)
	}
}

func TestEnvSnapshotIsImmutableDuringRun(t *testing.T) {
	const key = "TEST_RUN_ENV_SNAPSHOT"
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "book.yml")
	if err := os.WriteFile(path, []byte(`
desc: Tests for environment variable snapshot.
steps:
  -
    bind:
      changed: setenv('`+key+`', 'changed')
  -
    test: env.`+key+` == 'at-run-start'
  -
    include:
      path: included.yml
`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "included.yml"), []byte(`
desc: Tests for environment variable snapshot of included runbook.
steps:
  -
    test: env.`+key+` == 'at-run-start'
`), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(key, "at-run-start")
	o, err := New(Book(path), Scopes(scope.AllowReadParent), Func("setenv", func(k, v string) bool {
		os.Setenv(k, v)
		return true
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := o.Run(ctx); err != nil {
		t.Error(err)
	}
	// The process environment itself is changed, only the snapshot is kept.
	if got := os.Getenv(key); got != "changed" {
		t.Errorf("got %v\nwant %v", got, "changed")
	}
}
