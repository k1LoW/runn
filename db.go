package runn

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/xo/dburl"
)

type dbRunner struct {
	client   *sql.DB
	operator *operator
}

type dbQuery struct {
	stmt string
}

func newDBRunner(dsn string, o *operator) (*dbRunner, error) {
	db, err := dburl.Open(dsn)
	if err != nil {
		return nil, err
	}
	return &dbRunner{
		client:   db,
		operator: o,
	}, nil
}

func (rnr *dbRunner) Run(ctx context.Context, q *dbQuery) error {
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
	rnr.operator.store.steps = append(rnr.operator.store.steps, map[string]interface{}{
		"rows": rows,
	})
	return nil
}
