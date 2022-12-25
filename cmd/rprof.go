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
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/k1LoW/runn"
	"github.com/k1LoW/stopw"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var units = []string{"ns", "us", "ms", "s", "m"}
var sorts = []string{"elapsed", "started-at", "stopped-at"}

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
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT})
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetAutoFormatHeaders(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("-")
		table.SetHeaderLine(false)
		table.SetBorder(false)

		r := []row{}
		rr, err := appendBreakdown(s, 0, flags.ProfileDepth)
		if err != nil {
			return err
		}
		r = append(r, rr...)

		switch flags.ProfileSort {
		case "elapsed":
			sort.SliceStable(r, func(i, j int) bool {
				return r[i].elapsed < r[j].elapsed
			})
		case "started-at":
			sort.SliceStable(r, func(i, j int) bool {
				return r[i].startedAt.UnixNano() < r[j].startedAt.UnixNano()
			})
		case "stopped-at":
			sort.SliceStable(r, func(i, j int) bool {
				return r[i].stoppedAt.UnixNano() < r[j].stoppedAt.UnixNano()
			})
		default:
			if flags.ProfileSort != "" {
				return fmt.Errorf("invalid sort option: %s", flags.ProfileSort)
			}
		}

		d := make([][]string, len(r))
		for _, rr := range r {
			var id string
			switch rr.id.Type {
			case runn.IDTypeRunbook:
				id = fmt.Sprintf("%srunbook[%s](%s)", strings.Repeat("  ", rr.depth), rr.id.Desc, runn.ShortenPath(rr.id.RunbookPath))
			case runn.IDTypeStep:
				key := rr.id.StepRunnerKey
				if key == "" {
					key = string(rr.id.StepRunnerType)
				}
				id = fmt.Sprintf("%ssteps[%s].%s", strings.Repeat("  ", rr.depth), rr.id.StepKey, key)
			case runn.IDTypeBeforeFunc:
				id = fmt.Sprintf("%sbeforeFunc[%d]", strings.Repeat("  ", rr.depth), rr.id.FuncIndex)
			case runn.IDTypeAfterFunc:
				id = fmt.Sprintf("%safterFunc[%d]", strings.Repeat("  ", rr.depth), rr.id.FuncIndex)
			default:
				return fmt.Errorf("invalid runID type: %s", rr.id.Type)
			}
			d = append(d, []string{id, parseDuration(rr.elapsed)})
		}

		if flags.ProfileSort == "" {
			d = append(d, []string{"[total]", parseDuration(s.Elapsed)})
		}

		table.AppendBulk(d)
		table.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rprofCmd)
	rprofCmd.Flags().IntVarP(&flags.ProfileDepth, "depth", "", 4, flags.Usage("ProfileDepth"))
	rprofCmd.Flags().StringVarP(&flags.ProfileUnit, "unit", "", "ms", fmt.Sprintf(`time unit ("%s")`, strings.Join(units, `","`)))
	rprofCmd.Flags().StringVarP(&flags.ProfileSort, "sort", "", "", fmt.Sprintf(`sort order ("%s")`, strings.Join(sorts, `","`)))
}

type row struct {
	id        runn.ID
	elapsed   time.Duration
	startedAt time.Time
	stoppedAt time.Time
	depth     int
}

func appendBreakdown(p *stopw.Span, d, maxd int) ([]row, error) {
	if d > maxd {
		return nil, nil
	}
	rr := []row{}
	for _, s := range p.Breakdown {
		b, err := json.Marshal(s.ID)
		if err != nil {
			return nil, err
		}
		var runID runn.ID
		if err := json.Unmarshal(b, &runID); err != nil {
			return nil, err
		}
		rr = append(rr, row{runID, s.Elapsed, s.StartedAt, s.StoppedAt, d})
		rrr, err := appendBreakdown(s, d+1, maxd)
		if err != nil {
			return nil, err
		}
		rr = append(rr, rrr...)
	}
	return rr, nil
}

func parseDuration(d time.Duration) string {
	switch flags.ProfileUnit {
	case "ns":
		return fmt.Sprintf("%d%s", d, flags.ProfileUnit)
	case "us":
		return fmt.Sprintf("%.2f%s", float64(d)/float64(time.Microsecond), flags.ProfileUnit)
	case "ms":
		return fmt.Sprintf("%.2f%s", float64(d)/float64(time.Millisecond), flags.ProfileUnit)
	case "s":
		return fmt.Sprintf("%.2f%s", float64(d)/float64(time.Second), flags.ProfileUnit)
	case "m":
		return fmt.Sprintf("%.2f%s", float64(d)/float64(time.Minute), flags.ProfileUnit)
	default:
		return fmt.Sprintf("%dns", d)
	}
}
