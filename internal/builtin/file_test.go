package builtin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFile(t *testing.T) {
	tmpDir := t.TempDir()

	textContent := "Test text file content"
	textFilePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(textFilePath, []byte(textContent), 0644); err != nil {
		t.Fatalf("Failed to create text file: %v", err)
	}

	jsonContent := map[string]any{
		"name":    "test",
		"value":   123,
		"enabled": true,
	}
	jsonBytes, err := json.Marshal(jsonContent)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	jsonFilePath := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(jsonFilePath, jsonBytes, 0644); err != nil {
		t.Fatalf("Failed to create JSON file: %v", err)
	}

	dirPath := filepath.Join(tmpDir, "testdir")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
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
			want:    textContent,
			wantErr: false,
		},
		{
			name:    "Read text file (file:// scheme)",
			root:    tmpDir,
			path:    "file://test.txt",
			want:    textContent,
			wantErr: false,
		},
		{
			name:    "Read text file (absolute path)",
			root:    "/tmp", // not used
			path:    textFilePath,
			want:    textContent,
			wantErr: false,
		},
		{
			name: "Read JSON file (json:// scheme)",
			root: tmpDir,
			path: "json://test.json",
			want: map[string]any{
				"name":    "test",
				"value":   float64(123),
				"enabled": true,
			},
			wantErr: false,
		},
		{
			name:    "Read binary file (binary:// scheme)",
			root:    tmpDir,
			path:    "binary://test.txt",
			want:    []byte(textContent),
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
			name:    "Try to read a directory",
			root:    tmpDir,
			path:    "testdir",
			want:    nil,
			wantErr: true,
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := File(tt.root)
			got, err := fn(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("File() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
