package testutil

import (
	"strings"
	"testing"
)

func TestSQLite(t *testing.T) {
	_, dsn := SQLite(t)
	if !strings.Contains(dsn, "moderncsqlite") {
		t.Errorf("got %v\n", dsn)
	}
}
