package runn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/araddon/dateparse"
	_ "github.com/go-sql-driver/mysql"
	"github.com/goccy/go-json"
	"github.com/golang-sql/sqlexp"
	"github.com/golang-sql/sqlexp/nest"
	_ "github.com/googleapis/go-sql-spanner"
	"github.com/k1LoW/donegroup"
	_ "github.com/lib/pq"
	"github.com/xo/dburl"
	"modernc.org/sqlite"
)

const (
	dbStoreLastInsertIDKey = "last_insert_id"
	dbStoreRowsAffectedKey = "rows_affected"
	dbStoreRowsKey         = "rows"
)

type Querier interface {
	sqlexp.Querier
}

type TxQuerier interface {
	nest.Querier
	BeginTx(ctx context.Context, opts *nest.TxOptions) (*nest.Tx, error)
}

type dbRunner struct {
	name      string
	dsn       string
	client    TxQuerier
	hostRules hostRules
	trace     *bool
}

type dbQuery struct {
	stmt  string
	trace *bool
}

type DBResponse struct {
	LastInsertID int64
	RowsAffected int64
	Columns      []string
	Rows         []map[string]any
}

func newDBRunner(name, dsn string) (*dbRunner, error) {
	_, err := dburl.Parse(dsn)
	if err != nil {
		return nil, err
	}
	return &dbRunner{
		name: name,
		dsn:  dsn,
	}, nil
}

var dsnRep = strings.NewReplacer("sqlite://", "moderncsqlite://", "sqlite3://", "moderncsqlite://", "sq://", "moderncsqlite://")
var spannerInvalidatonKeyCounter uint64 = 0

func normalizeDSN(dsn string) string {
	if !contains(sql.Drivers(), "sqlite3") { // sqlite3 => github.com/mattn/go-sqlite3
		return dsnRep.Replace(dsn)
	}
	return dsn
}

func (rnr *dbRunner) Run(ctx context.Context, s *step) error {
	o := s.parent
	e, err := o.expandBeforeRecord(s.dbQuery)
	if err != nil {
		return err
	}
	q, ok := e.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid query: %v", e)
	}
	query, err := parseDBQuery(q)
	if err != nil {
		return fmt.Errorf("invalid query: %v %w", q, err)
	}
	if err := rnr.run(ctx, query, s); err != nil {
		return err
	}
	return nil
}

func (rnr *dbRunner) Close() error {
	if rnr.client == nil {
		return nil
	}
	if ndb, ok := rnr.client.(*nest.DB); ok {
		if db := ndb.DB(); db != nil {
			rnr.client = nil
			return db.Close()
		}
	}
	return nil
}

func (rnr *dbRunner) Renew() error {
	// If the DSN is in-memory, keep connection ( do not renew )
	if rnr.dsn == "sqlite3://:memory:" {
		return nil
	}
	if rnr.client != nil && rnr.dsn == "" {
		return errors.New("DB runners created with the runn.DBRunner option cannot be renewed") //nostyle:errorstrings
	}
	if err := rnr.Close(); err != nil {
		return err
	}
	return nil
}

func (rnr *dbRunner) run(ctx context.Context, q *dbQuery, s *step) error {
	o := s.parent
	if rnr.client == nil {
		if len(rnr.hostRules) > 0 {
			rnr.dsn = rnr.hostRules.replaceDSN(rnr.dsn)
		}
		nx, err := connectDB(rnr.dsn)
		if err != nil {
			return err
		}
		rnr.client = nx
		donegroup.Cleanup(ctx, func() error {
			return rnr.Renew()
		})
	}
	stmts := separateStmt(q.stmt)
	out := map[string]any{}
	tx, err := rnr.client.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	// Override trace
	switch {
	case q.trace == nil && rnr.trace == nil:
		q.trace = &o.trace
	case q.trace == nil && rnr.trace != nil:
		q.trace = rnr.trace
	}
	tc, err := q.generateTraceStmtComment(s)
	if err != nil {
		return err
	}
	for _, stmt := range stmts {
		stmt = stmt + tc // add trace comment
		o.capturers.captureDBStatement(rnr.name, stmt)
		err := func() error {
			if !isSELECTStmt(stmt) {
				// exec
				r, err := tx.ExecContext(ctx, stmt)
				if err != nil {
					return err
				}
				if isCommentOnlyStmt(stmt) {
					o.capturers.captureDBResponse(rnr.name, &DBResponse{})
					return nil
				}
				id, _ := r.LastInsertId()
				a, _ := r.RowsAffected()
				out = map[string]any{
					string(dbStoreLastInsertIDKey): id,
					string(dbStoreRowsAffectedKey): a,
				}

				o.capturers.captureDBResponse(rnr.name, &DBResponse{
					LastInsertID: id,
					RowsAffected: a,
				})

				return nil
			}

			// query
			var rows []map[string]any
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
				row := map[string]any{}
				vals := make([]any, len(columns))
				valsp := make([]any, len(columns))
				for i := range columns {
					valsp[i] = &vals[i]
				}
				if err := r.Scan(valsp...); err != nil {
					return err
				}
				for i, c := range columns {
					t := strings.ToUpper(types[i].DatabaseTypeName())
					switch v := vals[i].(type) {
					case []byte:
						s := string(v)
						switch {
						case strings.Contains(t, "TEXT") || strings.Contains(t, "CHAR") || t == "ENUM" || t == "TIME":
							row[c] = s
						case t == "DECIMAL" || t == "FLOAT" || t == "DOUBLE": // MySQL: NUMERIC = DECIMAL ( FLOAT, DOUBLE maybe not used )
							num, err := strconv.ParseFloat(s, 64) //nostyle:repetition
							if err != nil {
								return fmt.Errorf("invalid column: evaluated %s, but got %s(%v): %w", c, t, s, err)
							}
							row[c] = num
						case t == "DATE" || t == "TIMESTAMP" || t == "DATETIME": // MySQL(SSH port fowarding)
							d, err := dateparse.ParseStrict(s)
							if err != nil {
								return fmt.Errorf("invalid column: evaluated %s, but got %s(%v): %w", c, t, s, err)
							}
							row[c] = d
						case t == "JSONB": // PostgreSQL JSONB
							var jsonColumn map[string]any
							err = json.Unmarshal(v, &jsonColumn)
							if err != nil {
								return fmt.Errorf("invalid column: evaluated %s, but got %s(%v): %w", c, t, s, err)
							}
							row[c] = jsonColumn
						default: // MySQL: BOOLEAN = TINYINT
							num, err := strconv.Atoi(s) //nostyle:repetition
							if err != nil {
								return fmt.Errorf("invalid column: evaluated %s, but got %s(%v): %w", c, t, s, err)
							}
							row[c] = num
						}
					case string:
						switch {
						case t == "JSON": // Sqlite JSON
							var jsonColumn map[string]any
							err = json.Unmarshal([]byte(v), &jsonColumn)
							if err != nil {
								return fmt.Errorf("invalid column: evaluated %s, but got %s(%v): %w", c, t, v, err)
							}
							row[c] = jsonColumn
						default:
							row[c] = v
						}
					case float32: // MySQL: FLOAT
						s := strconv.FormatFloat(float64(v), 'f', -1, 32)
						num, err := strconv.ParseFloat(s, 64) //nostyle:repetition
						if err != nil {
							return fmt.Errorf("invalid column: evaluated %s, but got %s(%v): %w", c, t, v, err)
						}
						row[c] = num
					case float64: // MySQL: DOUBLE
						row[c] = v
					default:
						row[c] = v
					}
				}
				rows = append(rows, row)
			}
			if err := r.Err(); err != nil {
				return err
			}

			o.capturers.captureDBResponse(rnr.name, &DBResponse{
				Columns: columns,
				Rows:    rows,
			})

			out = map[string]any{
				string(dbStoreRowsKey): rows,
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
	o.record(out)
	return nil
}

func (q *dbQuery) generateTraceStmtComment(s *step) (string, error) {
	if q.trace == nil || !*q.trace {
		return "", nil
	}
	// Generate trace
	t := newTrace(s)
	// Trace structure to json
	tj, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	// Generate trace comment
	return fmt.Sprintf(" /* %s */", string(tj)), nil
}

func connectDB(dsn string) (TxQuerier, error) {
	var (
		db  *sql.DB
		err error
	)
	if strings.HasPrefix(dsn, "sp://") || strings.HasPrefix(dsn, "spanner://") {
		// NOTE: go-sql-spanner trys to reuse the connection internally when the same DSN is specified.
		key := atomic.AddUint64(&spannerInvalidatonKeyCounter, 1)
		d := strings.Split(strings.Split(dsn, "://")[1], "/")
		db, err = sql.Open("spanner", fmt.Sprintf(`projects/%s/instances/%s/databases/%s;workaroundConnectionInvalidationKey=%d`, d[0], d[1], d[2], key))
	} else {
		db, err = dburl.Open(normalizeDSN(dsn))
	}
	if err != nil {
		return nil, err
	}
	nx, err := nestTx(db)
	if err != nil {
		return nil, err
	}
	return nx, nil
}

func nestTx(client Querier) (TxQuerier, error) {
	switch c := client.(type) {
	case *sql.DB:
		return nest.Wrap(c), nil
	case *sql.Tx:
		if c == nil {
			return nil, fmt.Errorf("invalid db client: %v", c)
		}
		var v reflect.Value = reflect.ValueOf(c).Elem()
		var psv reflect.Value = v.FieldByName("db").Elem()
		db := (*sql.DB)(unsafe.Pointer(psv.UnsafeAddr()))
		return nest.Wrap(db), nil
	default:
		return nil, fmt.Errorf("invalid db client: %v", c)
	}
}

// reInlineComment - regexp for inline comment.
var reInlineComment = regexp.MustCompile(`\/\*.*?\*\/`)

func isSELECTStmt(stmt string) bool {
	stmt = strings.ToUpper(stmt)
	if !strings.Contains(stmt, "SELECT") {
		return false
	}
	lines := strings.Split(stmt, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(reInlineComment.ReplaceAllString(line, ""))
		if strings.HasPrefix(line, "SELECT") {
			return true
		}
	}
	return false
}

func isCommentOnlyStmt(stmt string) bool {
	stmt = strings.ToUpper(stmt)
	lines := strings.Split(stmt, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(reInlineComment.ReplaceAllString(line, ""))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}
		return false
	}
	return true
}

func separateStmt(stmt string) []string {
	stmt = strings.Trim(stmt, " \n\r")
	if !strings.Contains(stmt, ";") {
		return []string{stmt}
	}
	var (
		stmts []string
		s     []rune
	)
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
				stmts = append(stmts, strings.Trim(string(s), " \n\r"))
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

func init() {
	if !contains(sql.Drivers(), "moderncsqlite") {
		sql.Register("moderncsqlite", &sqlite.Driver{})
	}
}
