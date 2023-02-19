package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/xo/dburl"
	"modernc.org/sqlite"
)

func SQLite(t *testing.T) (*sql.DB, string) {
	t.Helper()
	p, err := os.CreateTemp("", "tmp")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Remove(p.Name())
	})
	dsn := fmt.Sprintf("moderncsqlite://%s", p.Name())
	db, err := dburl.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	return db, dsn
}

func init() {
	if !contains(sql.Drivers(), "moderncsqlite") {
		sql.Register("moderncsqlite", &sqlite.Driver{})
	}
}

func contains(s []string, e string) bool {
	for _, v := range s {
		if e == v {
			return true
		}
	}
	return false
}
