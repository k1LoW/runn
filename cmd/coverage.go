/*
Copyright © 2023 Ken'ichiro Oyama <k1lowxb@gmail.com>

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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1LoW/runn"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// coverageCmd represents the coverage command
var coverageCmd = &cobra.Command{
	Use:   "coverage [PATH_PATTERN ...]",
	Short: "show coverage for paths/operations of OpenAPI spec and methods of protocol buffers",
	Long:  `show coverage for paths/operations of OpenAPI spec and methods of protocol buffers.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		opts, err := flgs.ToOpts()
		if err != nil {
			return err
		}
		pathp := strings.Join(args, string(filepath.ListSeparator))
		opts = append(opts, runn.LoadOnly())

		// setup cache dir
		if err := runn.SetCacheDir(flgs.CacheDir); err != nil {
			return err
		}
		defer func() {
			if !flgs.RetainCacheDir {
				_ = runn.RemoveCacheDir()
			}
		}()

		o, err := runn.Load(pathp, opts...)
		if err != nil {
			return err
		}

		cov, err := o.CollectCoverage(ctx)
		if err != nil {
			return err
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetAutoFormatHeaders(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetHeaderLine(false)
		table.SetNoWhiteSpace(true)
		table.SetBorder(false)
		for _, spec := range cov.Specs {
			var total, covered int
			for _, v := range spec.Coverages {
				total++
				if v > 0 {
					covered++
				}
			}
			persent := float64(covered) / float64(total) * 100
			table.Append([]string{spec.Key + "  ", fmt.Sprintf("%.1f%%", persent)})
		}
		table.Render()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(coverageCmd)
	coverageCmd.Flags().StringSliceVarP(&flgs.Vars, "var", "", []string{}, flgs.Usage("Vars"))
	coverageCmd.Flags().StringSliceVarP(&flgs.Runners, "runner", "", []string{}, flgs.Usage("Runners"))
	coverageCmd.Flags().StringSliceVarP(&flgs.Overlays, "overlay", "", []string{}, flgs.Usage("Overlays"))
	coverageCmd.Flags().StringSliceVarP(&flgs.Underlays, "underlay", "", []string{}, flgs.Usage("Underlays"))
	coverageCmd.Flags().StringVarP(&flgs.RunMatch, "run", "", "", flgs.Usage("RunMatch"))
	coverageCmd.Flags().StringVarP(&flgs.RunID, "id", "", "", flgs.Usage("RunID"))
	coverageCmd.Flags().BoolVarP(&flgs.SkipIncluded, "skip-included", "", false, flgs.Usage("SkipIncluded"))
	coverageCmd.Flags().StringVarP(&flgs.CacheDir, "cache-dir", "", "", flgs.Usage("CacheDir"))
	coverageCmd.Flags().BoolVarP(&flgs.RetainCacheDir, "retain-cache-dir", "", false, flgs.Usage("RetainCacheDir"))
}
