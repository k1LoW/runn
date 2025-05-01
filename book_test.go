package runn

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goccy/go-json"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/k1LoW/runn/internal/scope"
	_ "github.com/lib/pq"
)

func TestNew(t *testing.T) {
	tests := []struct {
		path    string
		wantErr bool
	}{
		{"testdata/book/book.yml", false},
		{"testdata/book/map.yml", false},
		{"testdata/notexist.yml", true},
	}
	for _, tt := range tests {
		o, err := New(Book(tt.path))
		if err != nil {
			if !tt.wantErr {
				t.Errorf("got %v", err)
			}
			continue
		}
		if tt.wantErr {
			t.Errorf("want err")
		}
		if want := 1; len(o.httpRunners) != want {
			t.Errorf("got %v\nwant %v", len(o.httpRunners), want)
		}
		if want := 1; len(o.dbRunners) != want {
			t.Errorf("got %v\nwant %v", len(o.dbRunners), want)
		}
		if want := 6; len(o.steps) != want {
			t.Errorf("got %v\nwant %v", len(o.steps), want)
		}
	}
}

func TestLoadBook(t *testing.T) {
	tests := []struct {
		path      string
		varsBytes []byte
	}{
		{
			"testdata/book/env.yml",
			[]byte(`{"number": 1, "string": "string", "object": {"property": "property"}, "array": [ {"property": "property"} ] }`),
		},
	}
	debug := false
	t.Setenv("DEBUG", strconv.FormatBool(debug))
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			bk, err := LoadBook(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if want := debug; bk.debug != want {
				t.Errorf("got %v\nwant %v", bk.debug, want)
			}
			if want := "5"; bk.intervalStr != want {
				t.Errorf("got %v\nwant %v", bk.intervalStr, want)
			}
			got := bk.vars
			var want map[string]any
			if err := json.Unmarshal(tt.varsBytes, &want); err != nil {
				panic(err)
			}
			if diff := cmp.Diff(got, want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		opts []Option
		want any
	}{
		{[]Option{}, url.QueryEscape},
		{[]Option{Debug(true)}, url.QueryEscape},
		{[]Option{Func("getEnv", os.Getenv)}, url.QueryEscape},
		{[]Option{Func("urlencode", os.Getenv)}, os.Getenv},
	}
	for _, tt := range tests {
		bk := newBook()
		if err := bk.applyOptions(tt.opts...); err != nil {
			t.Fatal(err)
		}

		got := bk.funcs["urlencode"]
		if reflect.ValueOf(got).Pointer() != reflect.ValueOf(tt.want).Pointer() {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}

func TestApplyOptionsWithScope(t *testing.T) {
	tests := []struct {
		opts       []Option
		readRemote bool
		wantErr    bool
	}{
		{[]Option{Book("testdata/book/book.yml")}, false, false},
		{[]Option{Book("github://k1LoW/runn/testdata/book/http.yml")}, false, true},
		{[]Option{Book("github://k1LoW/runn/testdata/book/http.yml")}, true, false},
		{[]Option{Scopes(scope.AllowReadRemote), Book("github://k1LoW/runn/testdata/book/http.yml")}, false, false},
		{[]Option{Book("github://k1LoW/runn/testdata/book/http.yml"), Scopes(scope.AllowReadRemote)}, false, false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if tt.readRemote {
				if err := scope.Set(scope.AllowReadRemote); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := scope.Set(scope.DenyReadRemote); err != nil {
					t.Fatal(err)
				}
			}
			bk := newBook()
			if err := bk.applyOptions(tt.opts...); err != nil {
				if !tt.wantErr {
					t.Errorf("got %v", err)
				}
				return
			}
			if tt.wantErr {
				t.Error("want err")
			}
		})
	}
}

func TestParseRunnerForHttpRunner(t *testing.T) {
	secureUrl, _ := url.Parse("https://example.com/")
	url, _ := url.Parse("http://example.com/")
	client := &http.Client{Timeout: time.Duration(30000000000)}
	tests := []struct {
		v    any
		want any
	}{
		{
			"https://example.com/",
			httpRunner{
				name:            "req",
				endpoint:        secureUrl,
				client:          client,
				validator:       &nopValidator{},
				traceHeaderName: DefaultTraceHeaderName,
			},
		},
		{
			"http://example.com/",
			httpRunner{
				name:            "req",
				endpoint:        url,
				client:          client,
				validator:       &nopValidator{},
				traceHeaderName: DefaultTraceHeaderName,
			},
		},
	}
	opts := []cmp.Option{
		cmp.AllowUnexported(httpRunner{}),
		cmpopts.IgnoreFields(http.Client{}, "Transport"),
	}

	for _, tt := range tests {
		bk := newBook()
		if err := bk.parseRunner("req", tt.v); err != nil {
			t.Fatal(err)
		}

		got := bk.httpRunners["req"]
		if diff := cmp.Diff(*got, tt.want, opts...); diff != "" {
			t.Error(diff)
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
	bk := newBook()
	if err := bk.applyOptions(opt); err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		if bk.funcs[tt.fn] == nil {
			t.Errorf("not exists: %s", tt.fn)
		}
	}
}
