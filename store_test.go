package runn

import (
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestToMap(t *testing.T) {
	li := 1
	tests := []struct {
		store        store
		wantExistKey []string
	}{
		{
			store{},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			store{
				useMap: true,
			},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			store{
				parentVars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			store{
				bindVars: map[string]any{
					"bind": "value",
				},
			},
			[]string{"env", "vars", "steps", "bind", "parent", "runn", "needs"},
		},
		{
			store{
				loopIndex: &li,
			},
			[]string{"env", "vars", "steps", "i", "parent", "runn", "needs"},
		},
		{
			store{
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
		got := tt.store.toMap()
		gotKeys := make([]string, 0, len(got))
		for k := range got {
			gotKeys = append(gotKeys, k)
		}
		if diff := cmp.Diff(gotKeys, tt.wantExistKey, trns); diff != "" {
			t.Error(diff)
		}
	}
}

func TestToNormalizedMap(t *testing.T) {
	li := 1
	tests := []struct {
		store        store
		wantExistKey []string
	}{
		{
			store{},
			[]string{"env", "vars", "steps", "runn"},
		},
		{
			store{
				steps: []map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps", "runn"},
		},
		{
			store{
				useMap: true,
			},
			[]string{"env", "vars", "steps", "runn"},
		},
		{
			store{
				parentVars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps", "runn"},
		},
		{
			store{
				bindVars: map[string]any{
					"bind": "value",
				},
			},
			[]string{"env", "vars", "steps", "bind", "runn"},
		},
		{
			store{
				loopIndex: &li,
			},
			[]string{"env", "vars", "steps", "i", "runn"},
		},
		{
			store{
				cookies: map[string]map[string]*http.Cookie{},
			},
			[]string{"env", "vars", "steps", "cookies", "runn"},
		},
	}
	trns := cmp.Transformer("Sort", func(in []string) []string {
		out := append([]string(nil), in...) // Copy input to avoid mutating it
		sort.Strings(out)
		return out
	})
	for _, tt := range tests {
		got := tt.store.toNormalizedMap()
		gotKeys := make([]string, 0, len(got))
		for k := range got {
			gotKeys = append(gotKeys, k)
		}
		if diff := cmp.Diff(gotKeys, tt.wantExistKey, trns); diff != "" {
			t.Error(diff)
		}
	}
}

func TestRecordToCookie(t *testing.T) {
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
		store   store
		cookies []*http.Cookie
		want    map[string]map[string]*http.Cookie
	}{
		{
			store{},
			[]*http.Cookie{},
			map[string]map[string]*http.Cookie{},
		},
		{
			store{},
			[]*http.Cookie{&cookie1},
			map[string]map[string]*http.Cookie{
				"example.com": {
					"key1": &cookie1,
				},
			},
		},
		{
			store{},
			[]*http.Cookie{&cookie1, &cookie2},
			map[string]map[string]*http.Cookie{
				"example.com": {
					"key1": &cookie1,
					"key2": &cookie2,
				},
			},
		},
		{
			store{},
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
			store{
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
		tt.store.recordToCookie(tt.cookies)
		got := tt.store.toMap()["cookies"]
		opts := []cmp.Option{
			cmp.AllowUnexported(store{}),
		}
		if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
			t.Error(diff)
		}
	}
}
