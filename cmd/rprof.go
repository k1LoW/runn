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

// rprofCmd represents the rprof command.
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

		var r []row
		rr, err := appendBreakdown(s, 0, flgs.ProfileDepth)
		if err != nil {
			return err
		}
		r = append(r, rr...)

		switch flgs.ProfileSort {
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
			if flgs.ProfileSort != "" {
				return fmt.Errorf("invalid sort option: %s", flgs.ProfileSort)
			}
		}

		d := make([][]string, len(r))
		for _, rr := range r {
			var id string
			switch rr.trail.Type {
			case runn.TrailTypeRunbook:
				id = fmt.Sprintf("%srunbook[%s](%s)", strings.Repeat("  ", rr.depth), rr.trail.Desc, runn.ShortenPath(rr.trail.RunbookPath))
			case runn.TrailTypeStep:
				key := rr.trail.StepRunnerKey
				if key == "" {
					key = string(rr.trail.StepRunnerType)
				}
				id = fmt.Sprintf("%ssteps[%s].%s", strings.Repeat("  ", rr.depth), rr.trail.StepKey, key)
			case runn.TrailTypeBeforeFunc:
				id = fmt.Sprintf("%sbeforeFunc[%d]", strings.Repeat("  ", rr.depth), *rr.trail.FuncIndex)
			case runn.TrailTypeAfterFunc:
				id = fmt.Sprintf("%safterFunc[%d]", strings.Repeat("  ", rr.depth), *rr.trail.FuncIndex)
			case runn.TrailTypeLoop:
				id = fmt.Sprintf("%sloop[%d]", strings.Repeat("  ", rr.depth), *rr.trail.LoopIndex)
			default:
				return fmt.Errorf("invalid trail type: %s", rr.trail.Type)
			}
			d = append(d, []string{id, parseDuration(rr.elapsed)})
		}

		if flgs.ProfileSort == "" {
			d = append(d, []string{"[total]", parseDuration(s.Elapsed())})
		}

		table.AppendBulk(d)
		table.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(rprofCmd)
	rprofCmd.Flags().IntVarP(&flgs.ProfileDepth, "depth", "", 4, flgs.Usage("ProfileDepth"))
	rprofCmd.Flags().StringVarP(&flgs.ProfileUnit, "unit", "", "ms", fmt.Sprintf(`time unit (%q)`, strings.Join(units, `","`)))
	rprofCmd.Flags().StringVarP(&flgs.ProfileSort, "sort", "", "", fmt.Sprintf(`sort order (%q)`, strings.Join(sorts, `","`)))
}

type row struct {
	trail     runn.Trail
	elapsed   time.Duration
	startedAt time.Time
	stoppedAt time.Time
	depth     int
}

func appendBreakdown(p *stopw.Span, d, maxd int) ([]row, error) {
	if d > maxd {
		return nil, nil
	}
	var rr []row
	for _, s := range p.Breakdown {
		b, err := json.Marshal(s.ID)
		if err != nil {
			return nil, err
		}
		var tr runn.Trail
		if err := json.Unmarshal(b, &tr); err != nil {
			return nil, err
		}
		rr = append(rr, row{tr, s.Elapsed(), s.StartedAt, s.StoppedAt, d})
		rrr, err := appendBreakdown(s, d+1, maxd)
		if err != nil {
			return nil, err
		}
		rr = append(rr, rrr...)
	}
	return rr, nil
}

func parseDuration(d time.Duration) string {
	switch flgs.ProfileUnit {
	case "ns":
		return fmt.Sprintf("%d%s", d, flgs.ProfileUnit)
	case "us":
		return fmt.Sprintf("%.2f%s", float64(d)/float64(time.Microsecond), flgs.ProfileUnit)
	case "ms":
		return fmt.Sprintf("%.2f%s", float64(d)/float64(time.Millisecond), flgs.ProfileUnit)
	case "s":
		return fmt.Sprintf("%.2f%s", float64(d)/float64(time.Second), flgs.ProfileUnit)
	case "m":
		return fmt.Sprintf("%.2f%s", float64(d)/float64(time.Minute), flgs.ProfileUnit)
	default:
		return fmt.Sprintf("%dns", d)
	}
}
