package store

import (
	"fmt"
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/expr-lang/expr/parser"
	"github.com/google/go-cmp/cmp"
)

func TestStoreLatest(t *testing.T) {
	tests := []struct {
		name  string
		store Store
		want  map[string]any
	}{
		{
			"simple",
			Store{
				stepList: map[int]map[string]any{
					0: {"key": "zero"},
					1: {"key": "one"},
					2: {"key": "two"},
				},
			},
			map[string]any{
				"key": "two",
			},
		},
		{
			"no latest",
			Store{
				stepList: map[int]map[string]any{},
			},
			nil,
		},
		{
			"skipped",
			Store{
				stepList: map[int]map[string]any{
					1: {"key": "one"},
					4: {"key": "four"},
				},
			},
			map[string]any{
				"key": "four",
			},
		},
		{
			"simple map",
			Store{
				useMap: true,
				stepMap: map[string]map[string]any{
					"zero": {"key": "zero"},
					"one":  {"key": "one"},
					"two":  {"key": "two"},
				},
				stepMapKeys: []string{"zero", "one", "two"},
			},
			map[string]any{
				"key": "two",
			},
		},
		{
			"no latest map",
			Store{
				useMap:      true,
				stepMap:     map[string]map[string]any{},
				stepMapKeys: []string{"zero", "one", "two"},
			},
			nil,
		},
		{
			"skipped map",
			Store{
				useMap: true,
				stepMap: map[string]map[string]any{
					"one":  {"key": "one"},
					"four": {"key": "four"},
				},
				stepMapKeys: []string{"zero", "one", "two", "three", "four"},
			},
			map[string]any{
				"key": "four",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.store.Latest()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestStorePrevious(t *testing.T) {
	tests := []struct {
		name  string
		store Store
		want  map[string]any
	}{
		{
			"simple",
			Store{
				stepList: map[int]map[string]any{
					0: {"key": "zero"},
					1: {"key": "one"},
					2: {"key": "two"},
				},
			},
			map[string]any{
				"key": "one",
			},
		},
		{
			"no previous",
			Store{
				stepList: map[int]map[string]any{
					0: {"key": "zero"},
				},
			},
			nil,
		},
		{
			"skipped",
			Store{
				stepList: map[int]map[string]any{
					1: {"key": "one"},
					4: {"key": "four"},
				},
			},
			map[string]any{
				"key": "one",
			},
		},
		{
			"simple map",
			Store{
				useMap: true,
				stepMap: map[string]map[string]any{
					"zero": {"key": "zero"},
					"one":  {"key": "one"},
					"two":  {"key": "two"},
				},
				stepMapKeys: []string{"zero", "one", "two"},
			},
			map[string]any{
				"key": "one",
			},
		},
		{
			"no previous map",
			Store{
				useMap: true,
				stepMap: map[string]map[string]any{
					"zero": {"key": "zero"},
				},
				stepMapKeys: []string{"zero"},
			},
			nil,
		},
		{
			"skipped map",
			Store{
				useMap: true,
				stepMap: map[string]map[string]any{
					"one":  {"key": "one"},
					"four": {"key": "four"},
				},
				stepMapKeys: []string{"zero", "one", "two", "three", "four"},
			},
			map[string]any{
				"key": "one",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.store.Previous()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestToMap(t *testing.T) {
	li := 1
	tests := []struct {
		store        Store
		wantExistKey []string
	}{
		{
			Store{
				stepList: map[int]map[string]any{},
			},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			Store{
				stepList: map[int]map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			Store{
				useMap: true,
			},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			Store{
				parentVars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			Store{
				bindVars: map[string]any{
					"bind": "value",
				},
			},
			[]string{"env", "vars", "steps", "bind", "parent", "runn", "needs"},
		},
		{
			Store{
				loopIndex: &li,
			},
			[]string{"env", "vars", "steps", "i", "parent", "runn", "needs"},
		},
		{
			Store{
				cookies: map[string]map[string]*http.Cookie{},
			},
			[]string{"env", "vars", "steps", "cookies", "parent", "runn", "needs"},
		},
	}
	trns := cmp.Transformer("Sort", func(in []string) []string {
		out := append([]string(nil), in...) // Copy input to avoid mutating it
		sort.Strings(out)
		return out
	})
	for _, tt := range tests {
		got := tt.store.ToMap()
		gotKeys := make([]string, 0, len(got))
		for k := range got {
			gotKeys = append(gotKeys, k)
		}
		if diff := cmp.Diff(gotKeys, tt.wantExistKey, trns); diff != "" {
			t.Error(diff)
		}
	}
}

func TestToMapForIncludeRunner(t *testing.T) {
	li := 1
	tests := []struct {
		store        Store
		wantExistKey []string
	}{
		{
			Store{
				stepList: map[int]map[string]any{},
			},
			[]string{"env", "vars", "steps"},
		},
		{
			Store{
				stepList: map[int]map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps"},
		},
		{
			Store{
				useMap: true,
			},
			[]string{"env", "vars", "steps"},
		},
		{
			Store{
				parentVars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps"},
		},
		{
			Store{
				bindVars: map[string]any{
					"bind": "value",
				},
			},
			[]string{"env", "vars", "steps", "bind"},
		},
		{
			Store{
				loopIndex: &li,
			},
			[]string{"env", "vars", "steps", "i"},
		},
		{
			Store{
				cookies: map[string]map[string]*http.Cookie{},
			},
			[]string{"env", "vars", "steps", "cookies"},
		},
	}
	trns := cmp.Transformer("Sort", func(in []string) []string {
		out := append([]string(nil), in...) // Copy input to avoid mutating it
		sort.Strings(out)
		return out
	})
	for _, tt := range tests {
		got := tt.store.ToMapForIncludeRunner()
		gotKeys := make([]string, 0, len(got))
		for k := range got {
			gotKeys = append(gotKeys, k)
		}
		if diff := cmp.Diff(gotKeys, tt.wantExistKey, trns); diff != "" {
			t.Error(diff)
		}
	}
}

func TestRecordCookie(t *testing.T) {
	cookie1 := http.Cookie{
		Name:   "key1",
		Value:  "value1",
		Domain: "example.com",
	}
	cookie2 := http.Cookie{
		Name:   "key2",
		Value:  "value2",
		Domain: "example.com",
	}
	cookie3 := http.Cookie{
		Name:   "key3",
		Value:  "value3",
		Domain: "sub.example.com",
	}
	cookie4 := http.Cookie{
		Name:   "key1",
		Value:  "value4",
		Domain: "example.com",
	}
	cookie5 := http.Cookie{
		Name:    "key1",
		Value:   "value4",
		Domain:  "example.com",
		Expires: time.Now(),
	}
	tests := []struct {
		store   Store
		cookies []*http.Cookie
		want    map[string]map[string]*http.Cookie
	}{
		{
			Store{},
			[]*http.Cookie{},
			map[string]map[string]*http.Cookie{},
		},
		{
			Store{},
			[]*http.Cookie{&cookie1},
			map[string]map[string]*http.Cookie{
				"example.com": {
					"key1": &cookie1,
				},
			},
		},
		{
			Store{},
			[]*http.Cookie{&cookie1, &cookie2},
			map[string]map[string]*http.Cookie{
				"example.com": {
					"key1": &cookie1,
					"key2": &cookie2,
				},
			},
		},
		{
			Store{},
			[]*http.Cookie{&cookie1, &cookie2, &cookie3},
			map[string]map[string]*http.Cookie{
				"example.com": {
					"key1": &cookie1,
					"key2": &cookie2,
				},
				"sub.example.com": {
					"key3": &cookie3,
				},
			},
		},
		{
			Store{
				cookies: map[string]map[string]*http.Cookie{
					"example.com": {
						// Override
						"key1": &cookie4,
					},
				},
			},
			[]*http.Cookie{&cookie5},
			map[string]map[string]*http.Cookie{
				// Expire
				"example.com": {},
			},
		},
	}
	for _, tt := range tests {
		tt.store.RecordCookie(tt.cookies)
		got := tt.store.ToMap()["cookies"]
		opts := []cmp.Option{
			cmp.AllowUnexported(Store{}),
		}
		if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
			t.Error(diff)
		}
	}
}

func TestNodeToMap(t *testing.T) {
	v := "hello"
	tests := []struct {
		in    string
		store map[string]any
		want  map[string]any
	}{
		{
			"foo[3]",
			map[string]any{},
			map[string]any{
				"foo": map[any]any{
					3: v,
				},
			},
		},
		{
			"foo['hello']",
			map[string]any{},
			map[string]any{
				"foo": map[any]any{
					"hello": v,
				},
			},
		},
		{
			"foo['hello'][4]",
			map[string]any{},
			map[string]any{
				"foo": map[any]any{
					"hello": map[any]any{
						4: v,
					},
				},
			},
		},
		{
			"foo[5][4][3]",
			map[string]any{},
			map[string]any{
				"foo": map[any]any{
					5: map[any]any{
						4: map[any]any{
							3: v,
						},
					},
				},
			},
		},
		{
			"foo[key]",
			map[string]any{
				"key": "hello",
			},
			map[string]any{
				"foo": map[any]any{
					"hello": v,
				},
			},
		},
		{
			"foo[key][key2]",
			map[string]any{
				"key":  "hello",
				"key2": "hello2",
			},
			map[string]any{
				"foo": map[any]any{
					"hello": map[any]any{
						"hello2": v,
					},
				},
			},
		},
		{
			"foo[vars.key.key2]",
			map[string]any{
				"vars": map[any]any{
					"key": map[any]any{
						"key2": "hello",
					},
				},
			},
			map[string]any{
				"foo": map[any]any{
					"hello": v,
				},
			},
		},
		{
			"foo",
			map[string]any{},
			map[string]any{
				"foo": v,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			tr, err := parser.Parse(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			got, err := nodeToMap(tr.Node, v, tt.store)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMergeVars(t *testing.T) {
	tests := []struct {
		store map[string]any
		vars  map[string]any
		want  map[string]any
	}{
		{
			map[string]any{
				"key": "one",
			},
			map[string]any{
				"key": "two",
			},
			map[string]any{
				"key": "two",
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
				},
			},
			map[string]any{
				"parent": "two",
			},
			map[string]any{
				"parent": "two",
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
				},
			},
			map[string]any{
				"parent": []any{"two"},
			},
			map[string]any{
				"parent": []any{"two"},
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[string]any{
						"grandchild": "two",
					},
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child2": map[string]any{
						"grandchild": "three",
					},
					"child3": "three",
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child":  "one",
					"child2": map[string]any{"grandchild": "three"},
					"child3": "three",
				},
			},
		},
		{
			map[string]any{},
			map[string]any{
				"parent": map[any]any{
					0: "zero",
				},
			},
			map[string]any{
				"parent": map[any]any{
					0: "zero",
				},
			},
		},
		{
			map[string]any{
				"parent": map[any]any{
					0: "zero",
				},
			},
			map[string]any{
				"parent": map[any]any{
					1: "one",
				},
			},
			map[string]any{
				"parent": map[any]any{
					0: "zero",
					1: "one",
				},
			},
		},
		{
			map[string]any{
				"parent": map[any]any{
					"zero": "zero!",
				},
			},
			map[string]any{
				"parent": map[any]any{
					1: "one!",
				},
			},
			map[string]any{
				"parent": map[any]any{
					"zero": "zero!",
					1:      "one!",
				},
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[string]any{
						"grandchild": "two",
					},
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child2": map[any]any{
						"grandchild3": "three",
					},
					"child3": "three",
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[any]any{
						"grandchild":  "two",
						"grandchild3": "three",
					},
					"child3": "three",
				},
			},
		},
		{
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[any]any{
						"grandchild": "two",
					},
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child2": map[string]any{
						"grandchild3": "three",
					},
					"child3": "three",
				},
			},
			map[string]any{
				"parent": map[string]any{
					"child": "one",
					"child2": map[any]any{
						"grandchild":  "two",
						"grandchild3": "three",
					},
					"child3": "three",
				},
			},
		},
		{
			map[string]any{
				"parent": []any{
					"one",
				},
			},
			map[string]any{
				"parent": []any{
					"two",
				},
			},
			map[string]any{
				"parent": []any{
					"one",
					"two",
				},
			},
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			got := mergeVars(tt.store, tt.vars)
			if diff := cmp.Diff(got, tt.want, nil); diff != "" {
				t.Error(diff)
			}
		})
	}
}
