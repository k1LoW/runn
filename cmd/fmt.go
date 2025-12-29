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
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

var fmtWrite bool

// keyOrder defines the canonical order of keys in a runbook.
var keyOrder = []string{
	"desc",
	"labels",
	"needs",
	"runners",
	"vars",
	"secrets",
	"debug",
	"interval",
	"if",
	"skipTest",
	"loop",
	"concurrency",
	"force",
	"trace",
	"steps",
	"hostRules",
}

// fmtCmd represents the fmt command.
var fmtCmd = &cobra.Command{
	Use:     "fmt [PATH ...]",
	Short:   "format runbook YAML files",
	Long:    `format runbook YAML files with consistent style and key ordering.`,
	Aliases: []string{"format"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var hasError bool
		for _, path := range args {
			formatted, err := formatFile(path)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: %v\n", path, err)
				hasError = true
				continue
			}
			if fmtWrite {
				if err := os.WriteFile(path, formatted, 0o644); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "%s: %v\n", path, err)
					hasError = true
					continue
				}
			} else {
				io.WriteString(cmd.OutOrStdout(), string(formatted))
			}
		}
		if hasError {
			return errors.New("format failed")
		}
		return nil
	},
}

func formatFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse as MapSlice to preserve structure
	var ms yaml.MapSlice
	if err := yaml.Unmarshal(b, &ms); err != nil {
		return nil, err
	}

	// Reorder keys
	ordered := reorderKeys(ms)

	// Marshal with consistent formatting
	return yaml.Marshal(ordered)
}

func reorderKeys(ms yaml.MapSlice) yaml.MapSlice {
	// Create a map for quick lookup
	keyMap := make(map[string]yaml.MapItem)
	for _, item := range ms {
		if key, ok := item.Key.(string); ok {
			keyMap[key] = item
		}
	}

	// Build ordered slice
	var ordered yaml.MapSlice
	for _, key := range keyOrder {
		if item, ok := keyMap[key]; ok {
			ordered = append(ordered, item)
			delete(keyMap, key)
		}
	}

	// Append any remaining keys not in keyOrder (preserve original order)
	for _, item := range ms {
		if key, ok := item.Key.(string); ok {
			if _, exists := keyMap[key]; exists {
				ordered = append(ordered, item)
			}
		}
	}

	return ordered
}

func init() {
	rootCmd.AddCommand(fmtCmd)
	fmtCmd.Flags().BoolVarP(&fmtWrite, "write", "w", false, "write result to (source) file instead of stdout")
}
