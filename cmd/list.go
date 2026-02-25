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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/internal/fs"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

type listResultEntry struct {
	ID         string   `json:"id"`
	Desc       string   `json:"desc,omitempty"`
	Labels     []string `json:"labels,omitempty"`
	If         string   `json:"if,omitempty"`
	StepsCount int      `json:"steps_count"`
	Path       string   `json:"path"`
}

// listCmd represents the list command.
var listCmd = &cobra.Command{
	Use:     "list [PATH_PATTERN ...]",
	Short:   "list runbooks",
	Long:    `list runbooks.`,
	Aliases: []string{"ls"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pathp := strings.Join(args, string(filepath.ListSeparator))
		opts, err := flgs.ToOpts()
		if err != nil {
			return err
		}
		opts = append(opts, runn.LoadOnly())

		// setup cache dir
		if err := fs.SetCacheDir(flgs.CacheDir); err != nil {
			return err
		}
		defer func() {
			if !flgs.RetainCacheDir {
				_ = fs.RemoveCacheDir()
			}
		}()

		o, err := runn.Load(pathp, opts...)
		if err != nil {
			return err
		}
		selected, err := o.SelectedOperators()
		if err != nil {
			return err
		}

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			return err
		}

		if format == "json" {
			entries := make([]*listResultEntry, 0, len(selected))
			for _, oo := range selected {
				entries = append(entries, &listResultEntry{
					ID:         oo.ID(),
					Desc:       oo.Desc(),
					Labels:     oo.Labels(),
					If:         oo.If(),
					StepsCount: oo.NumberOfSteps(),
					Path:       oo.BookPath(),
				})
			}
			b, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(os.Stdout, string(b)); err != nil {
				return err
			}
			return nil
		}

		table := tablewriter.NewTable(os.Stdout,
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
				Borders: tw.BorderNone,
				Symbols: tw.NewSymbols(tw.StyleASCII),
				Settings: tw.Settings{
					Lines: tw.Lines{
						ShowTop:        tw.Off,
						ShowBottom:     tw.Off,
						ShowHeaderLine: tw.On,
						ShowFooterLine: tw.Off,
					},
					Separators: tw.Separators{
						ShowHeader:     tw.Off,
						ShowFooter:     tw.Off,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.Off,
					},
				},
			})),
			tablewriter.WithHeaderConfig(tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoFormat: tw.Off,
					Alignment:  tw.AlignLeft,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty},
				},
			}),
			tablewriter.WithRowConfig(tw.CellConfig{
				ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignLeft, tw.AlignLeft, tw.AlignRight, tw.AlignLeft},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty},
				},
			}),
		)
		table.Header([]string{"id:", "desc:", "if:", "steps:", "path"})
		for _, oo := range selected {
			id := oo.ID()
			if !flgs.Long {
				id = id[:7]
			}
			desc := oo.Desc()
			p := oo.BookPath()
			if !flgs.Long {
				p = fs.ShortenPath(p)
			}
			c := strconv.Itoa(oo.NumberOfSteps())
			ifCond := oo.If()
			if err := table.Append([]string{id, desc, ifCond, c, p}); err != nil {
				return err
			}
		}

		if err := table.Render(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolVarP(&flgs.Long, "long", "l", false, flgs.Usage("Long"))
	listCmd.Flags().BoolVarP(&flgs.SkipIncluded, "skip-included", "", false, flgs.Usage("SkipIncluded"))
	listCmd.Flags().StringSliceVarP(&flgs.Vars, "var", "", []string{}, flgs.Usage("Vars"))
	listCmd.Flags().StringSliceVarP(&flgs.Runners, "runner", "", []string{}, flgs.Usage("Runners"))
	listCmd.Flags().StringSliceVarP(&flgs.Overlays, "overlay", "", []string{}, flgs.Usage("Overlays"))
	listCmd.Flags().StringSliceVarP(&flgs.Underlays, "underlay", "", []string{}, flgs.Usage("Underlays"))
	listCmd.Flags().StringVarP(&flgs.RunMatch, "run", "", "", flgs.Usage("RunMatch"))
	listCmd.Flags().StringSliceVarP(&flgs.RunIDs, "id", "", []string{}, flgs.Usage("RunIDs"))
	listCmd.Flags().StringSliceVarP(&flgs.RunLabels, "label", "", []string{}, flgs.Usage("RunLabels"))
	listCmd.Flags().IntVarP(&flgs.Sample, "sample", "", 0, flgs.Usage("Sample"))
	listCmd.Flags().StringVarP(&flgs.Shuffle, "shuffle", "", "off", flgs.Usage("Shuffle"))
	listCmd.Flags().IntVarP(&flgs.Random, "random", "", 0, flgs.Usage("Random"))
	listCmd.Flags().IntVarP(&flgs.ShardIndex, "shard-index", "", 0, flgs.Usage("ShardIndex"))
	listCmd.Flags().IntVarP(&flgs.ShardN, "shard-n", "", 0, flgs.Usage("ShardN"))
	listCmd.Flags().StringVarP(&flgs.CacheDir, "cache-dir", "", "", flgs.Usage("CacheDir"))
	listCmd.Flags().BoolVarP(&flgs.RetainCacheDir, "retain-cache-dir", "", false, flgs.Usage("RetainCacheDir"))
	listCmd.Flags().StringVarP(&flgs.EnvFile, "env-file", "", "", flgs.Usage("EnvFile"))
	if err := listCmd.MarkFlagFilename("env-file"); err != nil {
		panic(err)
	}
	listCmd.Flags().StringVarP(&flgs.Format, "format", "", "", flgs.Usage("Format"))
}
