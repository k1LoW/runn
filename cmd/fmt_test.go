/*
Copyright © 2022 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFmtCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "format file to stdout",
			args:    []string{"fmt", "testdata/unformatted.yml"},
			wantErr: false,
		},
		{
			name:    "no arguments",
			args:    []string{"fmt"},
			wantErr: true,
		},
		{
			name:    "non-existent file",
			args:    []string{"fmt", "testdata/nonexistent.yml"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			rootCmd.SetOut(&bytes.Buffer{})
			rootCmd.SetErr(&bytes.Buffer{})

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("fmt command error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFmtCommandOutput(t *testing.T) {
	out := &bytes.Buffer{}
	rootCmd.SetArgs([]string{"fmt", "testdata/unformatted.yml"})
	rootCmd.SetOut(out)
	rootCmd.SetErr(&bytes.Buffer{})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the output contains the formatted YAML with correct key order
	output := out.String()

	// desc should come before runners
	descIdx := bytes.Index([]byte(output), []byte("desc:"))
	runnersIdx := bytes.Index([]byte(output), []byte("runners:"))
	stepsIdx := bytes.Index([]byte(output), []byte("steps:"))

	if descIdx == -1 || runnersIdx == -1 || stepsIdx == -1 {
		t.Fatalf("output missing expected keys: %s", output)
	}

	if descIdx > runnersIdx {
		t.Errorf("desc should come before runners in output")
	}
	if runnersIdx > stepsIdx {
		t.Errorf("runners should come before steps in output")
	}
}

func TestFmtCommandWriteOption(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yml")

	// Copy unformatted.yml to temp file
	content, err := os.ReadFile("testdata/unformatted.yml")
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	rootCmd.SetArgs([]string{"fmt", "--write", tmpFile})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read the formatted file
	formatted, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read formatted file: %v", err)
	}

	// Check key order
	descIdx := bytes.Index(formatted, []byte("desc:"))
	runnersIdx := bytes.Index(formatted, []byte("runners:"))
	stepsIdx := bytes.Index(formatted, []byte("steps:"))

	if descIdx == -1 || runnersIdx == -1 || stepsIdx == -1 {
		t.Fatalf("formatted file missing expected keys: %s", string(formatted))
	}

	if descIdx > runnersIdx {
		t.Errorf("desc should come before runners in formatted file")
	}
	if runnersIdx > stepsIdx {
		t.Errorf("runners should come before steps in formatted file")
	}
}
