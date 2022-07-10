package runn

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/k1LoW/grpcstub"
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
		{
			[]map[string]interface{}{},
			map[string]interface{}{"one": "01"},
			map[string]string{"path/{{ vars.one }}": "value"},
			map[string]interface{}{"path/01": "value"},
		},
		{
			[]map[string]interface{}{},
			map[string]interface{}{"year": 2022},
			map[string]string{"path?year={{ vars.year }}": "value"},
			map[string]interface{}{"path?year=2022": "value"},
		},
		{
			[]map[string]interface{}{},
			map[string]interface{}{"boolean": true},
			map[string]string{"boolean": "{{ vars.boolean }}"},
			map[string]interface{}{"boolean": true},
		},
		{
			[]map[string]interface{}{},
			map[string]interface{}{"nullable": nil},
			map[string]string{"nullable": "{{ vars.nullable }}"},
			map[string]interface{}{"nullable": nil},
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
		path     string
		RUNN_RUN string
		sample   int
		want     int
	}{
		{
			"testdata/book/**/*",
			"",
			0,
			func() int {
				e, _ := os.ReadDir("testdata/book/")
				return len(e)
			}(),
		},
		{"testdata/book/**/*", "initdb", 0, 1},
		{"testdata/book/**/*", "nonexistent", 0, 0},
		{"testdata/book/**/*", "", 3, 3},
		{
			"testdata/book/**/*",
			"",
			9999,
			func() int {
				e, _ := os.ReadDir("testdata/book/")
				return len(e)
			}(),
		},
	}
	for _, tt := range tests {
		t.Setenv("RUNN_RUN", tt.RUNN_RUN)
		opts := []Option{
			Runner("req", "https://api.github.com"),
			Runner("db", "sqlite://path/to/test.db"),
		}
		if tt.sample > 0 {
			opts = append(opts, RunSample(tt.sample))
		}
		ops, err := Load(tt.path, opts...)
		if err != nil {
			t.Fatal(err)
		}
		got := len(ops.ops)
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestSkipIncluded(t *testing.T) {
	tests := []struct {
		path         string
		skipIncluded bool
		want         int
	}{
		{"testdata/book/include_*", false, 3},
		{"testdata/book/include_*", true, 1},
	}
	for _, tt := range tests {
		ops, err := Load(tt.path, SkipIncluded(tt.skipIncluded), Runner("req", "https://api.github.com"), Runner("db", "sqlite://path/to/test.db"))
		if err != nil {
			t.Fatal(err)
		}
		got := len(ops.ops)
		if got != tt.want {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestSkipTest(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/skip_test.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(Book(tt.book))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestHookFuncTest(t *testing.T) {
	count := 0
	tests := []struct {
		book        string
		beforeFuncs []func() error
		afterFuncs  []func() error
		want        int
	}{
		{"testdata/book/skip_test.yml", nil, nil, 0},
		{
			"testdata/book/skip_test.yml",
			[]func() error{
				func() error {
					count += 3
					return nil
				},
				func() error {
					count = count * 2
					return nil
				},
			},
			[]func() error{
				func() error {
					count += 7
					return nil
				},
			},
			13,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		count = 0
		opts := []Option{
			Book(tt.book),
		}
		for _, fn := range tt.beforeFuncs {
			opts = append(opts, BeforeFunc(fn))
		}
		for _, fn := range tt.afterFuncs {
			opts = append(opts, AfterFunc(fn))
		}
		o, err := New(opts...)
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
		if count != tt.want {
			t.Errorf("got %v\nwant %v", count, tt.want)
		}
	}
}

func TestInclude(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/include_main.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(Book(tt.book), Func("upcase", strings.ToUpper))
		if err != nil {
			t.Fatal(err)
		}
		if err := o.Run(ctx); err != nil {
			t.Error(err)
		}
	}
}

func TestShard(t *testing.T) {
	tests := []struct {
		n int
	}{
		{2}, {3}, {4}, {5}, {6}, {7}, {11}, {13}, {17}, {999},
	}
	for _, tt := range tests {
		got := []*operator{}
		opts := []Option{
			Runner("req", "https://api.github.com"),
			Runner("db", "sqlite://path/to/test.db"),
		}
		all, err := Load("testdata/book/**/*", opts...)
		if err != nil {
			t.Fatal(err)
		}
		sortOperators(all.ops)
		want := all.ops
		for i := 0; i < tt.n; i++ {
			ops, err := Load("testdata/book/**/*", append(opts, RunShard(tt.n, i))...)
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, ops.ops...)
		}
		if len(got) != len(want) {
			t.Errorf("got %v\nwant %v", len(got), len(want))
		}
		sortOperators(got)
		allow := []interface{}{
			operator{}, httpRunner{}, dbRunner{}, grpcRunner{},
		}
		ignore := []interface{}{
			step{}, store{}, sql.DB{}, os.File{},
		}
		if diff := cmp.Diff(got, want, cmp.AllowUnexported(allow...), cmpopts.IgnoreUnexported(ignore...)); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestVars(t *testing.T) {
	tests := []struct {
		opts    []Option
		wantErr bool
	}{
		{
			[]Option{Book("testdata/book/vars.yml"), Var("token", "world")},
			false,
		},
		{
			[]Option{Book("testdata/book/vars.yml")},
			true,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		o, err := New(tt.opts...)
		if err != nil {
			t.Error(err)
		}
		if err := o.Run(ctx); err != nil {
			if !tt.wantErr {
				t.Errorf("got %v\n", err)
			}
			continue
		}
		if tt.wantErr {
			t.Error("want error")
		}
	}
}

func TestGrpc(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/grpc.yml"},
	}
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.book, func(t *testing.T) {
			t.Parallel()
			ts := grpcstub.NewServer(t, []string{}, "testdata/grpctest.proto")
			t.Cleanup(func() {
				ts.Close()
			})
			ts.Method("grpctest.GrpcTestService/Hello").
				Header("hello", "header").Trailer("hello", "trailer").
				ResponseString(`{"message":"hello", "num":32, "create_time":"2022-06-25T05:24:43.861872Z"}`)
			ts.Method("grpctest.GrpcTestService/ListHello").
				Header("listhello", "header").Trailer("listhello", "trailer").
				ResponseString(`{"message":"hello", "num":33, "create_time":"2022-06-25T05:24:43.861872Z"}`).
				ResponseString(`{"message":"hello", "num":34, "create_time":"2022-06-25T05:24:44.382783Z"}`)
			ts.Method("grpctest.GrpcTestService/MultiHello").
				Header("multihello", "header").Trailer("multihello", "trailer").
				ResponseString(`{"message":"hello", "num":35, "create_time":"2022-06-25T05:24:45.382783Z"}`)
			ts.Method("grpctest.GrpcTestService/HelloChat").Match(func(r *grpcstub.Request) bool {
				n, err := r.Message.Get("/name")
				if err != nil {
					return false
				}
				return n.(string) == "alice"
			}).Header("hellochat", "header").Trailer("hellochat", "trailer").
				ResponseString(`{"message":"hello", "num":34, "create_time":"2022-06-25T05:24:46.382783Z"}`)

			o, err := New(Book(tt.book), GrpcRunner("greq", ts.Conn()))
			if err != nil {
				t.Fatal(err)
			}
			if err := o.Run(ctx); err != nil {
				t.Error(err)
			}
		})
	}
}
