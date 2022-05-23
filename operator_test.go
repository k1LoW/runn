package runn

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/httpstub"
)

func TestExpand(t *testing.T) {
	tests := []struct {
		steps []map[string]interface{}
		vars  map[string]interface{}
		in    interface{}
		want  interface{}
	}{
		{
			[]map[string]interface{}{},
			map[string]interface{}{},
			map[string]string{"key": "val"},
			map[string]interface{}{"key": "val"},
		},
		{
			[]map[string]interface{}{},
			map[string]interface{}{"one": "ichi"},
			map[string]string{"key": "{{ vars.one }}"},
			map[string]interface{}{"key": "ichi"},
		},
		{
			[]map[string]interface{}{},
			map[string]interface{}{"one": "ichi"},
			map[string]string{"{{ vars.one }}": "val"},
			map[string]interface{}{"ichi": "val"},
		},
		{
			[]map[string]interface{}{},
			map[string]interface{}{"one": 1},
			map[string]string{"key": "{{ vars.one }}"},
			map[string]interface{}{"key": uint64(1)},
		},
		{
			[]map[string]interface{}{},
			map[string]interface{}{"one": 1},
			map[string]string{"key": "{{ vars.one + 1 }}"},
			map[string]interface{}{"key": uint64(2)},
		},
		{
			[]map[string]interface{}{},
			map[string]interface{}{"one": 1},
			map[string]string{"key": "{{ string(vars.one) }}"},
			map[string]interface{}{"key": "1"},
		},
	}
	for _, tt := range tests {
		o, err := New()
		if err != nil {
			t.Fatal(err)
		}
		o.store.steps = tt.steps
		o.store.vars = tt.vars

		got, err := o.expand(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(got, tt.want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestNewOption(t *testing.T) {
	tests := []struct {
		opts    []Option
		wantErr bool
	}{
		{
			[]Option{Book("testdata/book/book.yml"), Runner("db", "sqlite://path/to/test.db")},
			false,
		},
		{
			[]Option{Runner("db", "sqlite://path/to/test.db"), Book("testdata/book/book.yml")},
			false,
		},
		{
			[]Option{Book("testdata/book/notfound.yml")},
			true,
		},
		{
			[]Option{Runner("db", "unsupported://hostname")},
			true,
		},
		{
			[]Option{Runner("db", "sqlite://path/to/test.db"), HTTPRunner("db", "https://api.github.com", nil)},
			true,
		},
	}
	for _, tt := range tests {
		_, err := New(tt.opts...)
		got := (err != nil)
		if got != tt.wantErr {
			t.Errorf("got %v\nwant %v", got, tt.wantErr)
		}
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/db.yml"},
		{"testdata/book/only_if_included.yml"},
		{"testdata/book/if.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		func() {
			db, err := os.CreateTemp("", "tmp")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(db.Name())
			o, err := New(Book(tt.book), Runner("db", fmt.Sprintf("sqlite://%s", db.Name())))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		}()
	}
}

func TestRunAsT(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/db.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		func() {
			db, err := os.CreateTemp("", "tmp")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(db.Name())
			o, err := New(T(t), Book(tt.book), Runner("db", fmt.Sprintf("sqlite://%s", db.Name())))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		}()
	}
}

func TestRunUsingRetry(t *testing.T) {
	ts := httpstub.NewServer(t)
	counter := 0
	ts.Method(http.MethodGet).Handler(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fmt.Sprintf("%d", counter)))
		counter += 1
	})
	t.Cleanup(func() {
		ts.Close()
	})

	tests := []struct {
		book string
	}{
		{"testdata/book/retry.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(T(t), Book(tt.book), Runner("req", ts.Server().URL))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestRunUsingGitHubAPI(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skip("env GITHUB_TOKEN is not set")
	}
	tests := []struct {
		path string
	}{
		{"testdata/book/github.yml"},
		{"testdata/book/github_map.yml"},
	}
	for _, tt := range tests {
		ctx := context.Background()
		f, err := New(Book(tt.path))
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		path string
		want int
	}{
		{"testdata/book/*", 11},
		{"testdata/**/*", 12},
	}
	for _, tt := range tests {
		ops, err := Load(tt.path, Runner("req", "https://api.github.com"), Runner("db", "sqlite://path/to/test.db"))
		if err != nil {
			t.Fatal(err)
		}
		got := len(ops.ops)
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}
