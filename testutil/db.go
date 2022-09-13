package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/xo/dburl"
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
	db, err := dburl.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	return db, dsn
}
