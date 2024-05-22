package runn

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/runn/testutil"
)

func TestDBRunner(t *testing.T) {
	tests := []struct {
		stmt string
		want map[string]any
	}{
		{
			"SELECT 1",
			map[string]any{
				"rows": []map[string]any{
					{"1": int64(1)},
				},
			},
		},
		{
			"SELECT 1;SELECT 2;",
			map[string]any{
				"rows": []map[string]any{
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
			map[string]any{
				"last_insert_id": int64(1),
				"rows_affected":  int64(1),
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
			map[string]any{
				"rows": []map[string]any{
					{"count": int64(1)},
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
          updated NUMERIC,
		  info JSON
        );
INSERT INTO users (username, password, email, created, info) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'), '{
	"age": 20,
	"address": {
		"city": "Tokyo",
		"country": "Japan"
	}
}');
SELECT * FROM users;
`,
			map[string]any{
				"rows": []map[string]any{
					{
						"id":       int64(1),
						"username": "alice",
						"password": "passw0rd",
						"email":    "alice@example.com",
						"created":  "2017-12-05 00:00:00",
						"updated":  nil,
						"info": map[string]any{
							"age": float64(20),
							"address": map[string]any{
								"city":    "Tokyo",
								"country": "Japan",
							},
						},
					},
				},
			},
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.stmt, func(t *testing.T) {
			_, dsn := testutil.SQLite(t)
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			r, err := newDBRunner("db", dsn)
			if err != nil {
				t.Fatal(err)
			}
			s := newStep(0, "stepKey", o, nil)
			q := &dbQuery{stmt: tt.stmt}
			if err := r.run(ctx, q, s); err != nil {
				t.Error(err)
				return
			}
			got := o.store.steps[0]
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})

		t.Run(fmt.Sprintf("%s with Tx", tt.stmt), func(t *testing.T) {
			db, dsn := testutil.SQLite(t)
			o, err := New()
			if err != nil {
				t.Fatal(err)
			}
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
			})
			r, err := newDBRunner("db", dsn)
			if err != nil {
				t.Fatal(err)
			}
			nt, err := nestTx(tx)
			if err != nil {
				t.Fatal(err)
			}
			r.client = nt
			s := newStep(0, "stepKey", o, nil)
			q := &dbQuery{stmt: tt.stmt}
			if err := r.run(ctx, q, s); err != nil {
				t.Error(err)
				return
			}
			got := o.store.steps[0]
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
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
		{
			`CREATE TABLE users (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username TEXT UNIQUE NOT NULL,
          password TEXT NOT NULL,
          email TEXT UNIQUE NOT NULL,
          created NUMERIC NOT NULL,
          updated NUMERIC,
		  info JSON
        );
INSERT INTO users (username, password, email, created, info) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'), '{
	"age": 20,
	"address": {
		"city": "Tokyo",
		"country": "Japan"
	}
}');
SELECT * FROM users;
`,
			[]string{
				`CREATE TABLE users (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username TEXT UNIQUE NOT NULL,
          password TEXT NOT NULL,
          email TEXT UNIQUE NOT NULL,
          created NUMERIC NOT NULL,
          updated NUMERIC,
		  info JSON
        );`,
				`INSERT INTO users (username, password, email, created, info) VALUES ('alice', 'passw0rd', 'alice@example.com', datetime('2017-12-05'), '{
	"age": 20,
	"address": {
		"city": "Tokyo",
		"country": "Japan"
	}
}');`,
				"SELECT * FROM users;",
			},
		},
		{
			"SELECT 1\r",
			[]string{"SELECT 1"},
		},
		{
			"SELECT 1\n",
			[]string{"SELECT 1"},
		},
		{
			"SELECT 1;\rSELECT 2;\n",
			[]string{"SELECT 1;", "SELECT 2;"},
		},
	}
	for _, tt := range tests {
		got := separateStmt(tt.stmt)
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Error(diff)
		}
	}
}

func TestTraceStmtComment(t *testing.T) {
	tests := []struct {
		stmt string
	}{
		{
			"SELECT 1",
		},
		{
			"SELECT 1;SELECT 2;",
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
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.stmt, func(t *testing.T) {
			t.Run("runner with trace", func(t *testing.T) {
				_, dsn := testutil.SQLite(t)
				buf := new(bytes.Buffer)
				o, err := New(Capture(NewDebugger(buf)))
				if err != nil {
					t.Fatal(err)
				}
				trace := true
				r, err := newDBRunner("db", dsn)
				if err != nil {
					t.Fatal(err)
				}
				r.trace = &trace
				s := newStep(0, "stepKey", o, nil)
				q := &dbQuery{stmt: tt.stmt}
				if err := r.run(ctx, q, s); err != nil {
					t.Error(err)
					return
				}
				if !strings.Contains(buf.String(), "/* {") || !strings.Contains(buf.String(), "} */") {
					t.Errorf("trace comment not found: %s", buf.String())
				}
			})

			t.Run("query with trace", func(t *testing.T) {
				_, dsn := testutil.SQLite(t)
				buf := new(bytes.Buffer)
				o, err := New(Capture(NewDebugger(buf)))
				if err != nil {
					t.Fatal(err)
				}
				r, err := newDBRunner("db", dsn)
				if err != nil {
					t.Fatal(err)
				}
				trace := true
				s := newStep(0, "stepKey", o, nil)
				q := &dbQuery{stmt: tt.stmt, trace: &trace}
				if err := r.run(ctx, q, s); err != nil {
					t.Error(err)
					return
				}
				if !strings.Contains(buf.String(), "/* {") || !strings.Contains(buf.String(), "} */") {
					t.Errorf("trace comment not found: %s", buf.String())
				}
			})
		})
	}
}

func TestIsSELECTStmt(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"SELECT 1", true},
		{`
SELECT 1
`, true},
		{`--- comment
SELECT 1
`, true},
		{"/* comment */ SELECT 1", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got := isSELECTStmt(tt.in)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCommentOnlyStmt(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"SELECT 1", false},
		{`--- comment
SELECT 1
`, false},
		{"/* comment */ SELECT 1", false},
		{`--- comment
`, true},
		{"/* comment */ ", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got := isCommentOnlyStmt(tt.in)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
