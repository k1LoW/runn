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
	"os"
	"path/filepath"

	"github.com/k1LoW/runbk"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list [DIR]",
	Short:   "list books",
	Long:    `list books.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Desc", "Path"})
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoFormatHeaders(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("-")
		table.SetHeaderLine(true)
		table.SetBorder(false)

		for _, p := range args {
			f, err := os.Stat(p)
			if err != nil {
				return err
			}
			paths := []string{}
			if f.IsDir() {
				entries, err := os.ReadDir(p)
				if err != nil {
					return err
				}
				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					paths = append(paths, filepath.Join(p, e.Name()))
				}
			} else {
				paths = append(paths, p)
			}
			for _, p := range paths {
				b, err := runbk.LoadBookFile(p)
				if err == nil {
					desc := b.Desc
					if desc == "" {
						desc = runbk.NoDesc
					}
					table.Append([]string{desc, p})
				}
			}
		}

		table.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
