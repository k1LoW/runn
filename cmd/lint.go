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
	"os"

	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

// lintCmd represents the lint command.
var lintCmd = &cobra.Command{
	Use:   "lint [PATH ...]",
	Short: "lint runbook YAML files",
	Long:  `lint runbook YAML files for syntax errors.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var hasError bool
		for _, path := range args {
			if err := lintFile(path); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s: %v\n", path, err)
				hasError = true
			}
		}
		if hasError {
			return errors.New("lint failed")
		}
		return nil
	},
}

func lintFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
