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
			o, err := LoadBook(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if want := debug; o.debug != want {
				t.Errorf("got %v\nwant %v", o.debug, want)
			}
			if want := "5ms"; o.intervalStr != want {
				t.Errorf("got %v\nwant %v", o.intervalStr, want)
			}
			got := o.vars
			var want map[string]interface{}
			if err := json.Unmarshal(tt.varsBytes, &want); err != nil {
				panic(err)
			}
			if diff := cmp.Diff(got, want, nil); diff != "" {
				t.Errorf("%s", diff)
			}
		})
	}
}

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		opts []Option
		want interface{}
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
