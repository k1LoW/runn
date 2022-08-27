package runn

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDBRun(t *testing.T) {
	tests := []struct {
		stmt string
		want map[string]interface{}
	}{
		{
			"SELECT 1",
			map[string]interface{}{
				"rows": []map[string]interface{}{
					{"1": int64(1)},
				},
			},
		},
		{
			"SELECT 1;SELECT 2;",
			map[string]interface{}{
				"rows": []map[string]interface{}{
					{"2": int64(2)},
				},
			},
		},
		{
			`CREATE TABLE users (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username TEXT UNIQUE NOT NULL,
          password TEXT NOT NULL,
          email TEXT UNIQUE NOT NULL,
          created NUMERIC NOT NULL,
          updated NUMERIC
        );
INSERT INTO users (username, password, email, created) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'));`,
			map[string]interface{}{
				"last_insert_id": int64(1),
				"raws_affected":  int64(1),
			},
		},
		{
			`CREATE TABLE users (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username TEXT UNIQUE NOT NULL,
          password TEXT NOT NULL,
          email TEXT UNIQUE NOT NULL,
          created NUMERIC NOT NULL,
          updated NUMERIC
        );
INSERT INTO users (username, password, email, created) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'));
SELECT COUNT(*) AS count FROM users;
`,
			map[string]interface{}{
				"rows": []map[string]interface{}{
					{"count": int64(1)},
				},
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.stmt, func(t *testing.T) {
			db, err := os.CreateTemp("", "tmp")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(db.Name())
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			dsn := fmt.Sprintf("sqlite://%s", db.Name())
			r, err := newDBRunner("db", dsn, o)
			if err != nil {
				t.Fatal(err)
			}
			q := &dbQuery{stmt: tt.stmt}
			if err := r.Run(ctx, q); err != nil {
				t.Error(err)
				return
			}
			got := o.store.steps[0]
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func TestSeparateStmt(t *testing.T) {
	tests := []struct {
		stmt string
		want []string
	}{
		{
			"SELECT 1",
			[]string{"SELECT 1"},
		},
		{
			"SELECT 1;SELECT 2;",
			[]string{"SELECT 1;", "SELECT 2;"},
		},
		{
			`CREATE TABLE users (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username TEXT UNIQUE NOT NULL,
          password TEXT NOT NULL,
          email TEXT UNIQUE NOT NULL,
          created NUMERIC NOT NULL,
          updated NUMERIC
        );
INSERT INTO users (username, password, email, created) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'));`,
			[]string{
				`CREATE TABLE users (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username TEXT UNIQUE NOT NULL,
          password TEXT NOT NULL,
          email TEXT UNIQUE NOT NULL,
          created NUMERIC NOT NULL,
          updated NUMERIC
        );`,
				"INSERT INTO users (username, password, email, created) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'));",
			},
		},
		{
			`CREATE TABLE users (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username TEXT UNIQUE NOT NULL,
          password TEXT NOT NULL,
          email TEXT UNIQUE NOT NULL,
          created NUMERIC NOT NULL,
          updated NUMERIC
        );
INSERT INTO users (username, password, email, created) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'));
SELECT COUNT(*) AS count FROM users;
`,
			[]string{
				`CREATE TABLE users (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username TEXT UNIQUE NOT NULL,
          password TEXT NOT NULL,
          email TEXT UNIQUE NOT NULL,
          created NUMERIC NOT NULL,
          updated NUMERIC
        );`,
				"INSERT INTO users (username, password, email, created) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'));",
				"SELECT COUNT(*) AS count FROM users;",
			},
		},
	}
	for _, tt := range tests {
		got := separateStmt(tt.stmt)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}
