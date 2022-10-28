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
	"strings"

	"github.com/k1LoW/runn"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list [PATH_PATTERN ...]",
	Short:   "list runbooks",
	Long:    `list runbooks.`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Desc", "Path", "If"})
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoFormatHeaders(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("-")
		table.SetHeaderLine(true)
		table.SetBorder(false)

		pathp := strings.Join(args, string(filepath.ListSeparator))
		opts, err := flags.ToOpts()
		if err != nil {
			return err
		}
		o, err := runn.Load(pathp, opts...)
		if err != nil {
			return err
		}
		for _, oo := range o.Operators() {
			desc := oo.Desc()
			p := oo.BookPath()
			cond := oo.Cond()
			table.Append([]string{desc, p, cond})
		}

		table.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&flags.SkipIncluded, "skip-included", "", false, `skip running the included step by itself`)
	listCmd.Flags().StringSliceVarP(&flags.Vars, "var", "", []string{}, `set var to runbook ("key:value")`)
	listCmd.Flags().StringSliceVarP(&flags.Overlays, "overlay", "", []string{}, "overlay values on the runbook")
	listCmd.Flags().StringSliceVarP(&flags.Underlays, "underlay", "", []string{}, "lay values under the runbook")
	listCmd.Flags().IntVarP(&flags.Sample, "sample", "", 0, "run the specified number of runbooks at random")
	listCmd.Flags().StringVarP(&flags.Shuffle, "shuffle", "", "off", `randomize the order of running runbooks ("on","off",N)`)
}
