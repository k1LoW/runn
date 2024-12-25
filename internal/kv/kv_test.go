package kv

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestKV(t *testing.T) {
	tests := []struct {
		in any
	}{
		{nil},
		{"str"},
		{3},
		{4.5},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("set/get/del %v", tt.in), func(t *testing.T) {
			kv := New()
			kv.Set("key", tt.in)
			got := kv.Get("key")
			if diff := cmp.Diff(got, tt.in); diff != "" {
				t.Error(diff)
			}

			{
				kv.Del("key")
				got := kv.Get("key")
				if got != nil {
					t.Errorf("got %v, want %v", got, nil)
				}
			}
		})

		t.Run(fmt.Sprintf("set/get/clear %v", tt.in), func(t *testing.T) {
			kv := New()
			kv.Set("key", tt.in)
			got := kv.Get("key")
			if diff := cmp.Diff(got, tt.in); diff != "" {
				t.Error(diff)
			}

			{
				kv.Clear()
				got := kv.Get("key")
				if got != nil {
					t.Errorf("got %v, want %v", got, nil)
				}
			}
		})
	}
}
