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
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-json"
	"github.com/k1LoW/runn"
	"github.com/k1LoW/stopw"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// rprofCmd represents the rprof command
var rprofCmd = &cobra.Command{
	Use:     "rprof [PROFILE_PATH]",
	Short:   "read the runbook run profile",
	Long:    `read the runbook run profile.`,
	Aliases: []string{"rrprof", "rrrprof", "prof"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		b, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		var s *stopw.Span
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		s.Repair()
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoFormatHeaders(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("-")
		table.SetHeaderLine(false)
		table.SetBorder(false)

		d := [][]string{}
		dd, err := appendBreakdown(s, 0, flags.ProfileDepth)
		if err != nil {
			return err
		}
		d = append(d, dd...)

		d = append(d, []string{"", ""})
		d = append(d, []string{"total", s.Elapsed.String()})

		table.AppendBulk(d)
		table.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rprofCmd)
	rprofCmd.Flags().IntVarP(&flags.ProfileDepth, "depth", "", 4, "depth of profile")
}

func appendBreakdown(p *stopw.Span, d, maxd int) ([][]string, error) {
	if d > maxd {
		return nil, nil
	}
	dd := [][]string{}
	for _, s := range p.Breakdown {
		b, err := json.Marshal(s.ID)
		if err != nil {
			return nil, err
		}
		var (
			runID runn.ID
			id    string
		)
		if err := json.Unmarshal(b, &runID); err != nil {
			return nil, err
		}
		switch runID.Type {
		case runn.IDTypeRunbook:
			id = fmt.Sprintf("%srunbook[%s](%s)", strings.Repeat("  ", d), runID.Desc, runn.ShortenPath(runID.RunbookPath))
		case runn.IDTypeStep:
			key := runID.StepRunnerKey
			if key == "" {
				key = string(runID.StepRunnerType)
			}
			id = fmt.Sprintf("%ssteps[%s].%s", strings.Repeat("  ", d), runID.StepKey, key)
		case runn.IDTypeBeforeFunc:
			id = fmt.Sprintf("%sbeforeFunc[%d]", strings.Repeat("  ", d), runID.FuncIndex)
		case runn.IDTypeAfterFunc:
			id = fmt.Sprintf("%safterFunc[%d]", strings.Repeat("  ", d), runID.FuncIndex)
		default:
			return nil, fmt.Errorf("invalid runID type: %s", runID.Type)
		}
		dd = append(dd, []string{id, s.Elapsed.String()})
		ddd, err := appendBreakdown(s, d+1, maxd)
		if err != nil {
			return nil, err
		}
		dd = append(dd, ddd...)
	}
	return dd, nil
}
