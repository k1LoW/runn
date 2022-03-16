package runn

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
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
	for _, stmt := range stmts {
		if rnr.operator.debug {
			_, _ = fmt.Fprintf(os.Stderr, "-----START QUERY-----\n%s\n-----END QUERY-----\n", stmt)
		}
		err := func() error {
			if !strings.HasPrefix(strings.ToUpper(stmt), "SELECT") {
				// exec
				tx, err := rnr.client.Begin()
				if err != nil {
					return err
				}
				r, err := tx.ExecContext(ctx, stmt)
				if err != nil {
					return err
				}
				id, _ := r.LastInsertId()
				a, _ := r.RowsAffected()
				out = map[string]interface{}{
					"last_insert_id": id,
					"raws_affected":  a,
				}
				if err := tx.Commit(); err != nil {
					return err
				}
				return nil
			}

			// query
			rows := []map[string]interface{}{}
			r, err := rnr.client.QueryContext(ctx, stmt)
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
						if strings.Contains(t, "CHAR") || t == "TEXT" {
							row[c] = s
						} else {
							num, err := strconv.Atoi(s)
							if err != nil {
								return err
							}
							row[c] = num
						}
					default:
						row[c] = v
					}
				}
				rows = append(rows, row)
			}
			if err := r.Err(); err != nil {
				return err
			}
			if rnr.operator.debug {
				_, _ = fmt.Fprintln(os.Stderr, "-----START ROWS-----")
				table := tablewriter.NewWriter(os.Stderr)
				table.SetHeader(columns)
				table.SetAutoFormatHeaders(false)
				table.SetAutoWrapText(false)
				for _, r := range rows {
					row := make([]string, 0, len(columns))
					for _, c := range columns {
						row = append(row, fmt.Sprintf("%v", r[c]))
					}
					table.Append(row)
				}
				table.Render()
				c := len(rows)
				if c == 1 {
					_, _ = fmt.Fprintf(os.Stderr, "(%d row)\n", len(rows))
				} else {
					_, _ = fmt.Fprintf(os.Stderr, "(%d rows)\n", len(rows))
				}
				_, _ = fmt.Fprintln(os.Stderr, "-----END ROWS-----")
			}
			out = map[string]interface{}{
				"rows": rows,
			}
			return nil
		}()
		if err != nil {
			return err
		}
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
