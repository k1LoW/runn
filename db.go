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
	if rnr.operator.debug {
		_, _ = fmt.Fprintf(os.Stderr, "-----START QUERY-----\n%s-----END QUERY-----\n", q.stmt)
	}
	if !strings.HasPrefix(strings.ToUpper(q.stmt), "SELECT") {
		r, err := rnr.client.ExecContext(ctx, q.stmt)
		if err != nil {
			return err
		}
		id, err := r.LastInsertId()
		if err != nil {
			return err
		}
		a, err := r.RowsAffected()
		if err != nil {
			return err
		}
		rnr.operator.store.steps = append(rnr.operator.store.steps, map[string]interface{}{
			"last_insert_id": id,
			"raws_affected":  a,
		})
		return nil
	}
	rows := []map[string]interface{}{}
	r, err := rnr.client.QueryContext(ctx, q.stmt)
	if err != nil {
		return err
	}
	defer r.Close()

	columns, err := r.Columns()
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
				num, err := strconv.Atoi(s)
				if err == nil {
					row[c] = num
				} else {
					row[c] = s
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
	rnr.operator.store.steps = append(rnr.operator.store.steps, map[string]interface{}{
		"rows": rows,
	})
	return nil
}
