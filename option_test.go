package runn

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestOptionBook(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"testdata/book/book.yml", "Login and get projects."},
		{"testdata/book/db.yml", "Test using SQLite3"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			bk := newBook()
			opt := Book(tt.in)
			if err := opt(bk); err != nil {
				t.Fatal(err)
			}
			got := bk.desc
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestOptionOverlay(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		want    *book
		wantErr bool
	}{
		{
			"base",
			[]Option{
				Book("testdata/book/lay_1.yml"),
			},
			&book{
				desc:     "Test for layer(1)",
				runners:  map[string]any{"req": "https://example.com"},
				vars:     map[string]any{},
				rawSteps: []map[string]any{},
				path:     "testdata/book/lay_1.yml",
				httpRunners: map[string]*httpRunner{
					"req": {
						name:            "req",
						traceHeaderName: defaultTraceHeaderName,
					},
				},
				dbRunners:      map[string]*dbRunner{},
				grpcRunners:    map[string]*grpcRunner{},
				cdpRunners:     map[string]*cdpRunner{},
				sshRunners:     map[string]*sshRunner{},
				includeRunners: map[string]*includeRunner{},
				runnerErrs:     map[string]error{},
				useMap:         false,
			},
			false,
		},
		{
			"with overlay",
			[]Option{
				Book("testdata/book/lay_0.yml"),
				Overlay("testdata/book/lay_1.yml"),
			},
			&book{
				desc:    "Test for layer(1)",
				runners: map[string]any{"req": "https://example.com"},
				vars:    map[string]any{},
				rawSteps: []map[string]any{
					{"req": map[string]any{
						"/users": map[string]any{
							"get": map[string]any{
								"body": map[string]any{
									"application/json": nil,
								},
							},
						},
					}},
					{"req": map[string]any{
						"/users/1": map[string]any{
							"get": map[string]any{
								"body": map[string]any{
									"application/json": nil,
								},
							},
						},
					}},
				},
				stepKeys: []string{"get0", "get1"},
				path:     "testdata/book/lay_0.yml",
				httpRunners: map[string]*httpRunner{
					"req": {
						name:            "req",
						traceHeaderName: defaultTraceHeaderName,
					},
				},
				dbRunners:      map[string]*dbRunner{},
				grpcRunners:    map[string]*grpcRunner{},
				cdpRunners:     map[string]*cdpRunner{},
				sshRunners:     map[string]*sshRunner{},
				includeRunners: map[string]*includeRunner{},
				runnerErrs:     map[string]error{},
				useMap:         true,
			},
			false,
		},
		{
			"with overlay2",
			[]Option{
				Book("testdata/book/lay_0.yml"),
				Overlay("testdata/book/lay_1.yml"),
				Overlay("testdata/book/lay_2.yml"),
			},
			&book{
				desc: "Test for layer(2)",
				runners: map[string]any{
					"db":  "mysql://root:mypass@localhost:3306/testdb",
					"req": "https://example.com",
				},
				vars: map[string]any{},
				rawSteps: []map[string]any{
					{"req": map[string]any{
						"/users": map[string]any{
							"get": map[string]any{
								"body": map[string]any{
									"application/json": nil,
								},
							},
						},
					}},
					{"req": map[string]any{
						"/users/1": map[string]any{
							"get": map[string]any{
								"body": map[string]any{
									"application/json": nil,
								},
							},
						},
					}},
					{"db": map[string]any{
						"query": "SELECT * FROM users;",
					}},
				},
				stepKeys: []string{"get0", "get1", "db0"},
				path:     "testdata/book/lay_0.yml",
				httpRunners: map[string]*httpRunner{
					"req": {
						name:            "req",
						traceHeaderName: defaultTraceHeaderName,
					},
				},
				dbRunners: map[string]*dbRunner{
					"db": {name: "db", dsn: "mysql://root:mypass@localhost:3306/testdb"},
				},
				grpcRunners:    map[string]*grpcRunner{},
				cdpRunners:     map[string]*cdpRunner{},
				sshRunners:     map[string]*sshRunner{},
				includeRunners: map[string]*includeRunner{},
				runnerErrs:     map[string]error{},
				useMap:         true,
			},
			false,
		},
		{
			"overlay only",
			[]Option{
				Overlay("testdata/book/lay_0.yml"),
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newBook()
			for _, opt := range tt.opts {
				if err := opt(got); err != nil {
					if tt.wantErr {
						return
					}
				}
			}
			if tt.wantErr {
				t.Error("want error")
				return
			}
			opts := []cmp.Option{
				cmp.AllowUnexported(book{}, httpRunner{}, dbRunner{}),
				cmpopts.IgnoreFields(book{}, "funcs", "stdout", "stderr"),
				cmpopts.IgnoreFields(httpRunner{}, "endpoint", "client", "validator"),
				cmpopts.IgnoreFields(dbRunner{}, "client"),
			}
			if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestOptionUnderlay(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		want    *book
		wantErr bool
	}{
		{
			"base",
			[]Option{
				Book("testdata/book/lay_1.yml"),
			},
			&book{
				desc:     "Test for layer(1)",
				runners:  map[string]any{"req": "https://example.com"},
				vars:     map[string]any{},
				rawSteps: []map[string]any{},
				path:     "testdata/book/lay_1.yml",
				httpRunners: map[string]*httpRunner{
					"req": {
						name:            "req",
						traceHeaderName: defaultTraceHeaderName,
					},
				},
				dbRunners:      map[string]*dbRunner{},
				grpcRunners:    map[string]*grpcRunner{},
				cdpRunners:     map[string]*cdpRunner{},
				sshRunners:     map[string]*sshRunner{},
				includeRunners: map[string]*includeRunner{},
				runnerErrs:     map[string]error{},
				useMap:         false,
			},
			false,
		},
		{
			"with underlay",
			[]Option{
				Book("testdata/book/lay_0.yml"),
				Underlay("testdata/book/lay_1.yml"),
			},
			&book{
				desc:    "Test for layer(0)",
				runners: map[string]any{"req": "https://example.com"},
				vars:    map[string]any{},
				rawSteps: []map[string]any{
					{"req": map[string]any{
						"/users": map[string]any{
							"get": map[string]any{
								"body": map[string]any{
									"application/json": nil,
								},
							},
						},
					}},
					{"req": map[string]any{
						"/users/1": map[string]any{
							"get": map[string]any{
								"body": map[string]any{
									"application/json": nil,
								},
							},
						},
					}},
				},
				stepKeys: []string{"get0", "get1"},
				path:     "testdata/book/lay_0.yml",
				httpRunners: map[string]*httpRunner{
					"req": {
						name:            "req",
						traceHeaderName: defaultTraceHeaderName,
					},
				},
				dbRunners:      map[string]*dbRunner{},
				grpcRunners:    map[string]*grpcRunner{},
				cdpRunners:     map[string]*cdpRunner{},
				sshRunners:     map[string]*sshRunner{},
				includeRunners: map[string]*includeRunner{},
				runnerErrs:     map[string]error{},
				useMap:         true,
			},
			false,
		},
		{
			"with underlay2",
			[]Option{
				Book("testdata/book/lay_0.yml"),
				Underlay("testdata/book/lay_1.yml"),
				Underlay("testdata/book/lay_2.yml"),
			},
			&book{
				desc: "Test for layer(0)",
				runners: map[string]any{
					"db":  "mysql://root:mypass@localhost:3306/testdb",
					"req": "https://example.com",
				},
				vars: map[string]any{},
				rawSteps: []map[string]any{
					{"db": map[string]any{
						"query": "SELECT * FROM users;",
					}},
					{"req": map[string]any{
						"/users": map[string]any{
							"get": map[string]any{
								"body": map[string]any{
									"application/json": nil,
								},
							},
						},
					}},
					{"req": map[string]any{
						"/users/1": map[string]any{
							"get": map[string]any{
								"body": map[string]any{
									"application/json": nil,
								},
							},
						},
					}},
				},
				stepKeys: []string{"db0", "get0", "get1"},
				path:     "testdata/book/lay_0.yml",
				httpRunners: map[string]*httpRunner{
					"req": {
						name:            "req",
						traceHeaderName: defaultTraceHeaderName,
					},
				},
				dbRunners: map[string]*dbRunner{
					"db": {name: "db", dsn: "mysql://root:mypass@localhost:3306/testdb"},
				},
				grpcRunners:    map[string]*grpcRunner{},
				cdpRunners:     map[string]*cdpRunner{},
				sshRunners:     map[string]*sshRunner{},
				includeRunners: map[string]*includeRunner{},
				runnerErrs:     map[string]error{},
				useMap:         true,
			},
			false,
		},
		{
			"underlay only",
			[]Option{
				Underlay("testdata/book/lay_0.yml"),
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newBook()
			for _, opt := range tt.opts {
				if err := opt(got); err != nil {
					if tt.wantErr {
						return
					}
				}
			}
			if tt.wantErr {
				t.Error("want error")
				return
			}
			opts := []cmp.Option{
				cmp.AllowUnexported(book{}, httpRunner{}, dbRunner{}),
				cmpopts.IgnoreFields(book{}, "funcs", "stdout", "stderr"),
				cmpopts.IgnoreFields(httpRunner{}, "endpoint", "client", "validator"),
				cmpopts.IgnoreFields(dbRunner{}, "client"),
			}
			if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestOptionDesc(t *testing.T) {
	bk := newBook()

	opt := Desc("hello")
	if err := opt(bk); err != nil {
		t.Fatal(err)
	}

	got := bk.desc
	want := "hello"
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestOptionRunner(t *testing.T) {
	tests := []struct {
		name            string
		dsn             string
		opts            []httpRunnerOption
		wantHTTPRunners int
		wantDBRunners   int
		wantErrs        int
	}{
		{"req", "https://example.com/api/v1", nil, 1, 0, 0},
		{"db", "mysql://localhost/testdb", nil, 0, 1, 0},
		{"req", "https://example.com/api/v1", []httpRunnerOption{OpenAPI3("testdata/openapi3.yml")}, 1, 0, 0},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := Runner(tt.name, tt.dsn, tt.opts...)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}

		{
			got := len(bk.httpRunners)
			if got != tt.wantHTTPRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantHTTPRunners)
			}
		}

		{
			got := len(bk.dbRunners)
			if got != tt.wantDBRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantDBRunners)
			}
		}

		{
			got := len(bk.runnerErrs)
			if diff := cmp.Diff(got, tt.wantErrs, nil); diff != "" {
				t.Error(diff)
			}
		}
	}
}

func TestOptionHTTPRunner(t *testing.T) {
	tests := []struct {
		name            string
		endpoint        string
		client          *http.Client
		opts            []httpRunnerOption
		wantRunners     int
		wantHTTPRunners int
		wantErrs        int
	}{
		{"req", "https://api.example.com/v1", &http.Client{}, []httpRunnerOption{}, 0, 1, 0},
		{"req", "https://api.example.com/v1", &http.Client{}, []httpRunnerOption{HTTPTimeout("60s")}, 0, 1, 0},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := HTTPRunner(tt.name, tt.endpoint, tt.client, tt.opts...)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}

		{
			got := len(bk.runners)
			if got != tt.wantRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantRunners)
			}
		}

		{
			got := len(bk.httpRunners)
			if got != tt.wantHTTPRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantHTTPRunners)
			}
		}

		{
			got := len(bk.runnerErrs)
			if diff := cmp.Diff(got, tt.wantErrs, nil); diff != "" {
				t.Error(diff)
			}
		}
	}
}

func TestOptionHTTPRunnerWithHandler(t *testing.T) {
	tests := []struct {
		name            string
		handler         http.Handler
		opts            []httpRunnerOption
		wantRunners     int
		wantHTTPRunners int
		wantErrs        int
	}{
		{"req", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("hello k1LoW!")); err != nil {
				t.Fatal(err)
			}
		}), nil, 0, 1, 0},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := HTTPRunnerWithHandler(tt.name, tt.handler, tt.opts...)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}

		{
			got := len(bk.runners)
			if got != tt.wantRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantRunners)
			}
		}

		{
			got := len(bk.httpRunners)
			if got != tt.wantHTTPRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantHTTPRunners)
			}
		}

		{
			got := len(bk.runnerErrs)
			if diff := cmp.Diff(got, tt.wantErrs, nil); diff != "" {
				t.Error(diff)
			}
		}
	}
}

func TestOptionDBRunner(t *testing.T) {
	tests := []struct {
		name          string
		client        *sql.DB
		wantRunners   int
		wantDBRunners int
		wantErrs      int
	}{
		{"dbq", func() *sql.DB {
			db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/testdb")
			if err != nil {
				t.Fatal(err)
			}
			return db
		}(), 0, 1, 0},
		{"dbq", nil, 0, 1, 0},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := DBRunner(tt.name, tt.client)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}

		{
			got := len(bk.runners)
			if got != tt.wantRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantRunners)
			}
		}

		{
			got := len(bk.dbRunners)
			if got != tt.wantDBRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantDBRunners)
			}
		}

		{
			got := len(bk.runnerErrs)
			if diff := cmp.Diff(got, tt.wantErrs, nil); diff != "" {
				t.Error(diff)
			}
		}
	}
}

func TestOptionDBRunnerWithOptions(t *testing.T) {
	tests := []struct {
		name          string
		dsn           string
		opts          []dbRunnerOption
		wantRunners   int
		wantDBRunners int
	}{
		{"dbq", "mysql://username:password@localhost:3306/testdb", []dbRunnerOption{}, 0, 1},
		{"dbq", "mysql://username:password@localhost:3306/testdb", []dbRunnerOption{DBTrace(true)}, 0, 1},
	}

	for _, tt := range tests {
		bk := newBook()

		opt := DBRunnerWithOptions(tt.name, tt.dsn, tt.opts...)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}

		{
			got := len(bk.runners)
			if got != tt.wantRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantRunners)
			}
		}

		{
			got := len(bk.dbRunners)
			if got != tt.wantDBRunners {
				t.Errorf("got %v\nwant %v", got, tt.wantDBRunners)
			}
		}
	}
}

func TestOptionVar(t *testing.T) {
	tests := []struct {
		current map[string]any
		key     any
		value   any
		want    map[string]any
	}{
		{
			map[string]any{},
			"key", "value",
			map[string]any{
				"key": "value",
			},
		},
		{
			map[string]any{},
			"key", 3,
			map[string]any{
				"key": 3,
			},
		},
		{
			map[string]any{},
			[]string{"key"}, "value",
			map[string]any{
				"key": "value",
			},
		},
		{
			map[string]any{},
			[]string{"foo", "bar"}, "value",
			map[string]any{
				"foo": map[string]any{
					"bar": "value",
				},
			},
		},
		{
			map[string]any{
				"foo": map[string]any{
					"bar": "vaz",
				},
			},
			[]string{"foo", "bar"}, "value",
			map[string]any{
				"foo": map[string]any{
					"bar": "value",
				},
			},
		},
		{
			map[string]any{
				"foo": map[string]any{
					"bar": "baz",
					"qux": "quux",
				},
			},
			[]string{"foo", "bar"}, "value",
			map[string]any{
				"foo": map[string]any{
					"bar": "value",
					"qux": "quux",
				},
			},
		},
		{
			map[string]any{
				"foo": "xxx",
			},
			[]string{"foo", "bar"}, "value",
			map[string]any{
				"foo": map[string]any{
					"bar": "value",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v:%v", tt.key, tt.value), func(t *testing.T) {
			bk := newBook()
			bk.vars = tt.current
			opt := Var(tt.key, tt.value)
			if err := opt(bk); err != nil {
				t.Fatal(err)
			}
			got := bk.vars
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestOptionFunc(t *testing.T) {
	bk := newBook()

	if len(bk.vars) != 0 {
		t.Fatalf("got %v\nwant %v", len(bk.vars), 0)
	}

	opt := Func("sprintf", fmt.Sprintf)
	if err := opt(bk); err != nil {
		t.Fatal(err)
	}

	got, ok := bk.funcs["sprintf"].(func(string, ...any) string)
	if !ok {
		t.Fatalf("failed type assertion: %v", bk.funcs["sprintf"])
	}
	want := fmt.Sprintf
	if got("%s!", "hello") != want("%s!", "hello") {
		t.Errorf("got %v\nwant %v", got("%s!", "hello"), want("%s!", "hello"))
	}
}

func TestOptionIntarval(t *testing.T) {
	tests := []struct {
		d       time.Duration
		wantErr bool
	}{
		{1 * time.Second, false},
		{-1 * time.Second, true},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := Interval(tt.d)
		if err := opt(bk); err != nil {
			if !tt.wantErr {
				t.Errorf("got error %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
	}
}

func TestOptionGRPCNoTLS(t *testing.T) {
	tests := []struct {
		grpcNoTLS bool
		TLSs      []bool
		want      []bool
	}{
		{
			false,
			[]bool{true},
			[]bool{true},
		},
		{
			true,
			[]bool{true},
			[]bool{false},
		},
		{
			false,
			[]bool{true, false, true},
			[]bool{true, false, true},
		},
		{
			true,
			[]bool{true, false, true},
			[]bool{false, false, false},
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("grpcNoTLS=%v", tt.grpcNoTLS), func(t *testing.T) {
			opts := []Option{
				Book("testdata/book/vars.yml"),
				GRPCNoTLS(tt.grpcNoTLS),
			}
			for i, tls := range tt.TLSs {
				key := fmt.Sprintf("greq%d", i)
				opts = append(opts, GrpcRunnerWithOptions(key, "", TLS(tls)))
			}
			o, err := New(opts...)
			if err != nil {
				t.Error(err)
			}
			var got []bool
			for i := range tt.TLSs {
				key := fmt.Sprintf("greq%d", i)
				r, ok := o.grpcRunners[key]
				if !ok {
					t.Errorf("invalid key: %s", key)
				}
				got = append(got, *r.tls)
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestOptionRunMatch(t *testing.T) {
	tests := []struct {
		match string
	}{
		{""},
		{"regexp"},
	}
	for _, tt := range tests {
		bk := newBook()
		opt := RunMatch(tt.match)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}
		if bk.runMatch == nil {
			t.Error("bk.runMatch should not be nil")
		}
	}
}

func TestOptionRunSample(t *testing.T) {
	tests := []struct {
		sample  int
		wantErr bool
	}{
		{1, false},
		{3, false},
		{0, true},
		{-1, true},
	}
	for _, tt := range tests {
		bk := newBook()
		opt := RunSample(tt.sample)
		if err := opt(bk); err != nil {
			if !tt.wantErr {
				t.Errorf("got error %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		if bk.runSample != tt.sample {
			t.Errorf("got %v\nwant %v", bk.runSample, tt.sample)
		}
	}
}

func TestOptionRunRandom(t *testing.T) {
	tests := []struct {
		random  int
		wantErr bool
	}{
		{1, false},
		{3, false},
		{0, true},
		{-1, true},
	}
	for _, tt := range tests {
		bk := newBook()
		opt := RunRandom(tt.random)
		if err := opt(bk); err != nil {
			if !tt.wantErr {
				t.Errorf("got error %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		if bk.runRandom != tt.random {
			t.Errorf("got %v\nwant %v", bk.runRandom, tt.random)
		}
	}
}

func TestOptionRunShard(t *testing.T) {
	tests := []struct {
		n       int
		i       int
		wantErr bool
	}{
		{5, 1, false},
		{5, 1, false},
		{1, 0, false},
		{0, 0, true},
		{1, 1, true},
	}
	for _, tt := range tests {
		bk := newBook()
		opt := RunShard(tt.n, tt.i)
		if err := opt(bk); err != nil {
			if !tt.wantErr {
				t.Errorf("got error %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		if bk.runShardIndex != tt.i {
			t.Errorf("got %v\nwant %v", bk.runShardIndex, tt.i)
		}
		if bk.runShardN != tt.n {
			t.Errorf("got %v\nwant %v", bk.runShardN, tt.n)
		}
	}
}

func TestSetupBuiltinFunctions(t *testing.T) {
	tests := []struct {
		fn string
	}{
		{"url"},
		{"urlencode"},
		{"bool"},
		{"time"},
		{"compare"},
		{"diff"},
		{"pick"},
		{"intersect"},
		{"sprintf"},
		{"basename"},
		{"faker"},
	}
	opt := Func("sprintf", fmt.Sprintf)
	opts := setupBuiltinFunctions(opt)
	bk := newBook()
	for _, o := range opts {
		if err := o(bk); err != nil {
			t.Fatal(err)
		}
	}
	for _, tt := range tests {
		if bk.funcs[tt.fn] == nil {
			t.Errorf("not exists: %s", tt.fn)
		}
	}
}

func TestOptionNotFollowRedirect(t *testing.T) {
	tests := []struct {
		notFollowRedirect bool
		wantNil           bool
	}{
		{false, true},
		{true, false},
	}
	for _, tt := range tests {
		key := "req"
		t.Run(fmt.Sprintf("Runner notFollowRedirect:%v", tt.notFollowRedirect), func(t *testing.T) {
			bk := newBook()
			opt := Runner(key, "https://example.com", NotFollowRedirect(tt.notFollowRedirect))
			if err := opt(bk); err != nil {
				t.Error(err)
			}
			if bk.httpRunners[key] == nil {
				t.Error("got nil\nwant *httpRunner")
			}
			if tt.wantNil {
				if bk.httpRunners[key].client.CheckRedirect != nil {
					t.Error("got func\nwant nil")
				}
			} else {
				if bk.httpRunners[key].client.CheckRedirect == nil {
					t.Error("got nil\nwant func")
				}
			}
		})

		t.Run(fmt.Sprintf("HTTPRunner notFollowRedirect:%v", tt.notFollowRedirect), func(t *testing.T) {
			bk := newBook()
			opt := HTTPRunner(key, "https://example.com", http.DefaultClient, NotFollowRedirect(tt.notFollowRedirect))
			if err := opt(bk); err != nil {
				t.Error(err)
			}
			if bk.httpRunners[key] == nil {
				t.Error("got nil\nwant *httpRunner")
			}
			if tt.wantNil {
				if bk.httpRunners[key].client.CheckRedirect != nil {
					t.Error("got func\nwant nil")
				}
			} else {
				if bk.httpRunners[key].client.CheckRedirect == nil {
					t.Error("got nil\nwant func")
				}
			}
		})
	}
}

func TestOptionRunID(t *testing.T) {
	tests := []struct {
		ids  []string
		want []string
	}{
		{nil, nil},
		{[]string{"a"}, []string{"a"}},
		{[]string{""}, nil},
		{[]string{"a", "b"}, []string{"a", "b"}},
		{[]string{"b", "a"}, []string{"b", "a"}},
		{[]string{"a", "b,c"}, []string{"a", "b", "c"}},
		{[]string{"a\nb", "c"}, []string{"a", "b", "c"}},
		{[]string{"a\nb\nc\n"}, []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.ids), func(t *testing.T) {
			bk := newBook()
			opt := RunID(tt.ids...)
			if err := opt(bk); err != nil {
				t.Error(err)
			}
			if diff := cmp.Diff(tt.want, bk.runIDs); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestBuiltinFunctionBooks(t *testing.T) {
	tests := []struct {
		book    string
		wantErr bool
	}{
		{"testdata/book/builtin_pick.yml", false},
		{"testdata/book/builtin_omit.yml", false},
		{"testdata/book/builtin_merge.yml", false},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.book, func(t *testing.T) {
			t.Parallel()
			o, err := New(Book(tt.book))
			if err != nil {
				if !tt.wantErr {
					t.Errorf("got %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Errorf("want err")
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestOptionHostRules(t *testing.T) {
	tests := []struct {
		hostRules []string
		want      hostRules
		wantErr   bool
	}{
		{
			nil, nil, false,
		},
		{
			[]string{"example.com 127.0.0.1"}, hostRules{{host: "example.com", rule: "127.0.0.1"}}, false,
		},
		{
			[]string{"*.example.com 127.0.0.1:80"}, hostRules{{host: "*.example.com", rule: "127.0.0.1:80"}}, false,
		},
		{
			[]string{"example.com"}, nil, true,
		},
		{
			[]string{"example.com 127.0.0.1", "app.example.com 127.0.0.1"}, hostRules{{host: "example.com", rule: "127.0.0.1"}, {host: "app.example.com", rule: "127.0.0.1"}}, false,
		},
		{
			[]string{"example.com 127.0.0.1, app.example.com 127.0.0.1"}, hostRules{{host: "example.com", rule: "127.0.0.1"}, {host: "app.example.com", rule: "127.0.0.1"}}, false,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.hostRules), func(t *testing.T) {
			bk := newBook()
			opt := HostRules(tt.hostRules...)
			err := opt(bk)
			if (err != nil) != tt.wantErr {
				t.Errorf("got %v", err)
				return
			}
			opts := []cmp.Option{
				cmp.AllowUnexported(hostRule{}),
			}
			if diff := cmp.Diff(tt.want, bk.hostRulesFromOpts, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}
