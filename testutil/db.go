package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
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
	dsn := fmt.Sprintf("sqlite://%s", p.Name())
	db, err := dburl.Open(normalizeDSN(dsn))
	if err != nil {
		t.Fatal(err)
	}
	return db, dsn
}

var dsnRep = strings.NewReplacer("sqlite://", "moderncsqlite://", "sqlite3://", "moderncsqlite://", "sq://", "moderncsqlite://")

func normalizeDSN(dsn string) string {
	if !contains(sql.Drivers(), "sqlite3") { // sqlite3 => github.com/mattn/go-sqlite3
		return dsnRep.Replace(dsn)
	}
	return dsn
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
