package runn

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/runn/testutil"
	"github.com/tenntenn/golden"
)

func TestCoverage(t *testing.T) {
	tests := []struct {
		book string
	}{
		{"testdata/book/httpbin.yml"},
		{"testdata/book/grpc.yml"},
	}
	t.Setenv("DEBUG", "false")
	ctx := context.Background()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.book, func(t *testing.T) {
			t.Parallel()
			o, err := New(Book(tt.book))
			if err != nil {
				t.Fatal(err)
			}
			cov, err := o.collectCoverage(ctx)
			if err != nil {
				t.Fatal(err)
			}
			got, err := json.Marshal(cov)
			if err != nil {
				t.Fatal(err)
			}
			f := fmt.Sprintf("%s.coverage.json", filepath.Base(tt.book))
			if os.Getenv("UPDATE_GOLDEN") != "" {
				golden.Update(t, testutil.Testdata(), f, got)
				return
			}
			if diff := golden.Diff(t, testutil.Testdata(), f, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
