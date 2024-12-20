package runn

import (
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestStoreLatest(t *testing.T) {
	tests := []struct {
		name  string
		store store
		want  map[string]any
	}{
		{
			"simple",
			store{
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
			store{
				stepList: map[int]map[string]any{},
			},
			nil,
		},
		{
			"skipped",
			store{
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
			store{
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
			store{
				useMap:      true,
				stepMap:     map[string]map[string]any{},
				stepMapKeys: []string{"zero", "one", "two"},
			},
			nil,
		},
		{
			"skipped map",
			store{
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
			got := tt.store.latest()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestStorePrevious(t *testing.T) {
	tests := []struct {
		name  string
		store store
		want  map[string]any
	}{
		{
			"simple",
			store{
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
			store{
				stepList: map[int]map[string]any{
					0: {"key": "zero"},
				},
			},
			nil,
		},
		{
			"skipped",
			store{
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
			store{
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
			store{
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
			store{
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
			got := tt.store.previous()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestToMap(t *testing.T) {
	li := 1
	tests := []struct {
		store        store
		wantExistKey []string
	}{
		{
			store{
				stepList: map[int]map[string]any{},
			},
			[]string{"env", "vars", "steps", "parent", "runn", "needs"},
		},
		{
			store{
				stepList: map[int]map[string]any{},
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

func TestToMapForIncludeRunner(t *testing.T) {
	li := 1
	tests := []struct {
		store        store
		wantExistKey []string
	}{
		{
			store{
				stepList: map[int]map[string]any{},
			},
			[]string{"env", "vars", "steps"},
		},
		{
			store{
				stepList: map[int]map[string]any{},
				vars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps"},
		},
		{
			store{
				useMap: true,
			},
			[]string{"env", "vars", "steps"},
		},
		{
			store{
				parentVars: map[string]any{
					"key": "value",
				},
			},
			[]string{"env", "vars", "steps"},
		},
		{
			store{
				bindVars: map[string]any{
					"bind": "value",
				},
			},
			[]string{"env", "vars", "steps", "bind"},
		},
		{
			store{
				loopIndex: &li,
			},
			[]string{"env", "vars", "steps", "i"},
		},
		{
			store{
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
		got := tt.store.toMapForIncludeRunner()
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
		tt.store.recordCookie(tt.cookies)
		got := tt.store.toMap()["cookies"]
		opts := []cmp.Option{
			cmp.AllowUnexported(store{}),
		}
		if diff := cmp.Diff(got, tt.want, opts...); diff != "" {
			t.Error(diff)
		}
	}
}
