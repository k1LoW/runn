package runn

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
		want    string
		wantErr bool
	}{
		{
			"base",
			[]Option{
				Book("testdata/book/book.yml"),
			},
			"Login and get projects.",
			false,
		},
		{
			"with overlay",
			[]Option{
				Book("testdata/book/book.yml"),
				Overlay("testdata/book/db.yml"),
			},
			"Test using SQLite3",
			false,
		},
		{
			"with overlay2",
			[]Option{
				Book("testdata/book/book.yml"),
				Overlay("testdata/book/db.yml"),
				Overlay("testdata/book/dump.yml"),
			},
			"For dump test",
			false,
		},
		{
			"overlay only",
			[]Option{
				Overlay("testdata/book/book.yml"),
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bk := newBook()
			for _, opt := range tt.opts {
				if err := opt(bk); err != nil {
					if tt.wantErr {
						return
					}
				}
			}
			if tt.wantErr {
				t.Error("want error")
				return
			}
			got := bk.desc
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
			}
		})
	}
}

func TestOptionUnderlay(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		want    string
		wantErr bool
	}{
		{
			"base",
			[]Option{
				Book("testdata/book/book.yml"),
			},
			"Login and get projects.",
			false,
		},
		{
			"with underlay",
			[]Option{
				Book("testdata/book/book.yml"),
				Underlay("testdata/book/db.yml"),
			},
			"Login and get projects.",
			false,
		},
		{
			"with underlay2",
			[]Option{
				Book("testdata/book/book.yml"),
				Underlay("testdata/book/db.yml"),
				Underlay("testdata/book/dump.yml"),
			},
			"Login and get projects.",
			false,
		},
		{
			"underlay only",
			[]Option{
				Underlay("testdata/book/book.yml"),
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bk := newBook()
			for _, opt := range tt.opts {
				if err := opt(bk); err != nil {
					if tt.wantErr {
						return
					}
				}
			}
			if tt.wantErr {
				t.Error("want error")
				return
			}
			got := bk.desc
			if got != tt.want {
				t.Errorf("got %v\nwant %v", got, tt.want)
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
		{"req", "https://example.com/api/v1", []httpRunnerOption{OpenApi3("testdata/openapi3.yml")}, 1, 0, 0},
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
				t.Errorf("%s", diff)
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
				t.Errorf("%s", diff)
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
				t.Errorf("%s", diff)
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
		{"req", func() *sql.DB {
			db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/testdb")
			if err != nil {
				t.Fatal(err)
			}
			return db
		}(), 0, 1, 0},
		{"req", nil, 0, 1, 0},
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
				t.Errorf("%s", diff)
			}
		}
	}
}

func TestOptionVar(t *testing.T) {
	bk := newBook()

	if len(bk.vars) != 0 {
		t.Fatalf("got %v\nwant %v", len(bk.vars), 0)
	}

	opt := Var("key", "value")
	if err := opt(bk); err != nil {
		t.Fatal(err)
	}

	got, ok := bk.vars["key"].(string)
	if !ok {
		t.Fatalf("failed type assertion: %v", bk.vars["key"])
	}
	want := "value"
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
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

	got, ok := bk.funcs["sprintf"].(func(string, ...interface{}) string)
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
