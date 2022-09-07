package runn

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/xo/dburl"
)

type dbRunner struct {
	name     string
	client   *sql.DB
	operator *operator
}

type dbQuery struct {
	stmt string
}

type DBResponse struct {
	LastInsertID int64
	RowsAffected int64
	Columns      []string
	Rows         []map[string]interface{}
}

func newDBRunner(name, dsn string, o *operator) (*dbRunner, error) {
	db, err := dburl.Open(dsn)
	if err != nil {
		return nil, err
	}
	return &dbRunner{
		name:     name,
		client:   db,
		operator: o,
	}, nil
}

func (rnr *dbRunner) Run(ctx context.Context, q *dbQuery) error {
	stmts := separateStmt(q.stmt)
	out := map[string]interface{}{}
	tx, err := rnr.client.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	for _, stmt := range stmts {
		rnr.operator.capturers.captureDBStatement(stmt)
		err := func() error {
			if !strings.HasPrefix(strings.ToUpper(stmt), "SELECT") {
				// exec
				r, err := tx.ExecContext(ctx, stmt)
				if err != nil {
					return err
				}
				id, _ := r.LastInsertId()
				a, _ := r.RowsAffected()
				out = map[string]interface{}{
					"last_insert_id": id,
					"rows_affected":  a,
				}

				rnr.operator.capturers.captureDBResponse(&DBResponse{
					LastInsertID: id,
					RowsAffected: a,
				})

				return nil
			}

			// query
			rows := []map[string]interface{}{}
			r, err := tx.QueryContext(ctx, stmt)
			if err != nil {
				return err
			}
			defer r.Close()

			columns, err := r.Columns()
			if err != nil {
				return err
			}
			types, err := r.ColumnTypes()
			if err != nil {
				return err
			}
			for r.Next() {
				row := map[string]interface{}{}
				vals := make([]interface{}, len(columns))
				valsp := make([]interface{}, len(columns))
				for i := range columns {
					valsp[i] = &vals[i]
				}
				if err := r.Scan(valsp...); err != nil {
					return err
				}
				for i, c := range columns {
					switch v := vals[i].(type) {
					case []byte:
						s := string(v)
						t := strings.ToUpper(types[i].DatabaseTypeName())
						switch {
						case strings.Contains(t, "TEXT") || strings.Contains(t, "CHAR") || t == "TIME": // MySQL8: ENUM = CHAR
							row[c] = s
						case t == "DECIMAL" || t == "FLOAT" || t == "DOUBLE": // MySQL: NUMERIC = DECIMAL
							num, err := strconv.ParseFloat(s, 64)
							if err != nil {
								return fmt.Errorf("invalid column: evaluated %s, but got %s(%v): %w", c, t, s, err)
							}
							row[c] = num
						default: // MySQL: BOOLEAN = TINYINT
							num, err := strconv.Atoi(s)
							if err != nil {
								return fmt.Errorf("invalid column: evaluated %s, but got %s(%v): %w", c, t, s, err)
							}
							row[c] = num
						}
					default:
						// MySQL8: DATE, TIMESTAMP, DATETIME
						row[c] = v
					}
				}
				rows = append(rows, row)
			}
			if err := r.Err(); err != nil {
				return err
			}

			rnr.operator.capturers.captureDBResponse(&DBResponse{
				Columns: columns,
				Rows:    rows,
			})

			out = map[string]interface{}{
				"rows": rows,
			}
			return nil
		}()
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	rnr.operator.record(out)
	return nil
}

func separateStmt(stmt string) []string {
	if !strings.Contains(stmt, ";") {
		return []string{stmt}
	}
	stmts := []string{}
	s := []rune{}
	ins := false
	ind := false
	for _, c := range stmt {
		s = append(s, c)
		switch c {
		case '\'':
			ins = !ins
		case '"':
			ind = !ind
		case ';':
			if !ins && !ind {
				stmts = append(stmts, strings.Trim(string(s), " \n"))
				s = []rune{}
			}
		}
	}
	if len(s) > 0 {
		cutset := " \n\\n\"" // When I receive a multi-line query with `key: |`, I get an unexplained string at the end. Therefore, remove it as a workaround.
		l := strings.TrimRight(string(s), cutset)
		if len(l) > 0 {
			stmts = append(stmts, l)
		}
	}
	return stmts
}
