package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
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
		_ = os.Remove(p.Name()) //nolint:gosec
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
	if !slices.Contains(sql.Drivers(), "sqlite3") { // sqlite3 => github.com/mattn/go-sqlite3
		return dsnRep.Replace(dsn)
	}
	return dsn
}

func init() {
	if !slices.Contains(sql.Drivers(), "moderncsqlite") {
		sql.Register("moderncsqlite", &sqlite.Driver{})
	}
}
