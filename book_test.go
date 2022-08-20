package runn

import (
	"net/url"
	"os"
	"reflect"
	"strconv"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/goccy/go-json"
	"github.com/google/go-cmp/cmp"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func TestNew(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"testdata/book/book.yml"},
		{"testdata/book/map.yml"},
	}
	for _, tt := range tests {
		o, err := New(Book(tt.path))
		if err != nil {
			t.Fatal(err)
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
	os.Setenv("DEBUG", strconv.FormatBool(debug))
	for _, tt := range tests {
		o, err := LoadBook(tt.path)
		if err != nil {
			t.Fatal(err)
		}
		if want := debug; o.Debug != want {
			t.Errorf("got %v\nwant %v", o.Debug, want)
		}
		if want := "5"; o.Interval != want {
			t.Errorf("got %v\nwant %v", o.Interval, want)
		}
		got := o.Vars
		var want map[string]interface{}
		if err := json.Unmarshal(tt.varsBytes, &want); err != nil {
			panic(err)
		}
		if diff := cmp.Diff(got, want, nil); diff != "" {
			t.Errorf("%s", diff)
		}
	}
}

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		opts []Option
		want interface{}
	}{
		{[]Option{}, url.QueryEscape},
		{[]Option{Debug(true)}, url.QueryEscape},
		{[]Option{Func("gtEnv", os.Getenv)}, url.QueryEscape},
		{[]Option{Func("urlencode", os.Getenv)}, os.Getenv},
	}
	for _, tt := range tests {
		bk := newBook()
		if err := bk.ApplyOptions(tt.opts...); err != nil {
			t.Fatal(err)
		}

		got := bk.Funcs["urlencode"]
		if reflect.DeepEqual(got, tt.want) {
			t.Errorf("got %v\nwant %v", got, tt.want)
		}
	}
}
