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
	}
	for _, tt := range tests {
		bk := newBook()
		opt := Book(tt.in)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}
		got := bk.Desc
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestOptionDesc(t *testing.T) {
	bk := newBook()

	opt := Desc("hello")
	if err := opt(bk); err != nil {
		t.Fatal(err)
	}

	got := bk.Desc
	want := "hello"
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestOptionRunner(t *testing.T) {
	tests := []struct {
		name            string
		dsn             string
		opts            []RunnerOption
		wantRunners     int
		wantHTTPRunners int
		wantDBRunners   int
		wantErrs        int
	}{
		{"req", "https://example.com/api/v1", nil, 1, 0, 0, 0},
		{"db", "mysql://localhost/testdb", nil, 1, 0, 0, 0},
		{"req", "https://example.com/api/v1", []RunnerOption{OpenApi3("testdata/openapi3.yml")}, 0, 1, 0, 0},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := Runner(tt.name, tt.dsn, tt.opts...)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}

		{
			got := len(bk.Runners)
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
		opts            []RunnerOption
		wantRunners     int
		wantHTTPRunners int
		wantErrs        int
	}{
		{"req", "https://api.example.com/v1", &http.Client{}, []RunnerOption{}, 0, 1, 0},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := HTTPRunner(tt.name, tt.endpoint, tt.client, tt.opts...)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}

		{
			got := len(bk.Runners)
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
		opts            []RunnerOption
		wantRunners     int
		wantHTTPRunners int
		wantErrs        int
	}{
		{"req", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("hello k1LoW!"))
		}), nil, 0, 1, 0},
	}
	for _, tt := range tests {
		bk := newBook()

		opt := HTTPRunnerWithHandler(tt.name, tt.handler, tt.opts...)
		if err := opt(bk); err != nil {
			t.Fatal(err)
		}

		{
			got := len(bk.Runners)
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
			db, _ := sql.Open("mysql", "username:password@tcp(localhost:3306)/testdb")
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
			got := len(bk.Runners)
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

	if len(bk.Vars) != 0 {
		t.Fatalf("got %v\nwant %v", len(bk.Vars), 0)
	}

	opt := Var("key", "value")
	if err := opt(bk); err != nil {
		t.Fatal(err)
	}

	got := bk.Vars["key"].(string)
	want := "value"
	if got != want {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

func TestOptionFunc(t *testing.T) {
	bk := newBook()

	if len(bk.Vars) != 0 {
		t.Fatalf("got %v\nwant %v", len(bk.Vars), 0)
	}

	opt := Func("sprintf", fmt.Sprintf)
	if err := opt(bk); err != nil {
		t.Fatal(err)
	}

	got := bk.Funcs["sprintf"].(func(string, ...interface{}) string)
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

func TestOptionRunPart(t *testing.T) {
	tests := []struct {
		i       int
		n       int
		wantErr bool
	}{
		{1, 5, false},
		{0, 5, false},
		{0, 0, true},
		{1, 1, true},
	}
	for _, tt := range tests {
		bk := newBook()
		opt := RunPart(tt.n, tt.i)
		if err := opt(bk); err != nil {
			if !tt.wantErr {
				t.Errorf("got error %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
		if bk.runPartIndex != tt.i {
			t.Errorf("got %v\nwant %v", bk.runPartIndex, tt.i)
		}
		if bk.runPartN != tt.n {
			t.Errorf("got %v\nwant %v", bk.runPartN, tt.n)
		}
	}
}
