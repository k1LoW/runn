package builtin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/k1LoW/runn/internal/scope"
)

func TestFile(t *testing.T) {
	tmpDir := t.TempDir()

	testContent := "Test text file content"
	testFilePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFilePath, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	tests := []struct {
		name    string
		root    string
		path    string
		want    any
		wantErr bool
	}{
		{
			name:    "Read text file (relative path)",
			root:    tmpDir,
			path:    "test.txt",
			want:    testContent,
			wantErr: false,
		},
		{
			name:    "Read text file (file:// scheme)",
			root:    tmpDir,
			path:    "file://test.txt",
			want:    testContent,
			wantErr: false,
		},
		{
			name:    "Read text file (absolute path)",
			root:    "/tmp", // not used
			path:    testFilePath,
			want:    testContent,
			wantErr: false,
		},
		{
			name:    "Read non-existent file",
			root:    tmpDir,
			path:    "notexist.txt",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "Use unsupported scheme",
			root:    tmpDir,
			path:    "yaml://test.txt",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Specify only scheme part in path",
			root:    tmpDir,
			path:    "file://",
			want:    nil,
			wantErr: true,
		},
	}

	scope.Set(scope.AllowReadParent)
	t.Cleanup(func() {
		scope.Set(scope.DenyReadParent)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := File(tt.root)
			got, err := fn(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("File() error = %v, wantErr %v", err, tt.wantErr) //nostyle:errorstrings
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
