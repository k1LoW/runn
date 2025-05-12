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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/internal/fs"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var sortByMethod = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}

// coverageCmd represents the coverage command.
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

		cov, err := o.CollectCoverage(ctx)
		if err != nil {
			return err
		}

		if flgs.Format == "json" {
			b, err := json.MarshalIndent(cov, "", "  ")
			if err != nil {
				return err
			}
			_, _ = fmt.Println(string(b))
			return nil
		}

		table := tablewriter.NewTable(os.Stdout,
			tablewriter.WithTrimSpace(tw.Off),
			tablewriter.WithRenderer(renderer.NewColorized(renderer.ColorizedConfig{
				Borders: tw.BorderNone,
				Symbols: tw.NewSymbols(tw.StyleASCII),
				Header: renderer.Tint{
					FG: renderer.Colors{color.Bold},
					BG: renderer.Colors{color.Bold},
				},
				Column: renderer.Tint{
					FG: renderer.Colors{color.FgWhite},
					BG: renderer.Colors{color.FgWhite},
				},
				Settings: tw.Settings{
					Separators: tw.Separators{
						ShowHeader:     tw.On,
						ShowFooter:     tw.Off,
						BetweenRows:    tw.Off,
						BetweenColumns: tw.Off,
					},
				},
			})),
			tablewriter.WithHeaderConfig(tw.CellConfig{
				Formatting: tw.CellFormatting{
					AutoFormat: false,
					Alignment:  tw.AlignLeft,
				},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty},
				},
			}),
			tablewriter.WithRowConfig(tw.CellConfig{
				ColumnAligns: []tw.Align{tw.AlignLeft, tw.AlignRight},
				Padding: tw.CellPadding{
					Global: tw.Padding{Left: tw.Space, Right: tw.Space, Top: tw.Empty, Bottom: tw.Empty},
				},
			}),
		)
		ct := "Coverage"
		if flgs.Long {
			ct = "Coverage/Count"
		}
		table.Header([]string{"Spec", ct})
		var (
			coverages      [][]string
			total, covered int
		)
		for _, spec := range cov.Specs {
			var t, c int
			for _, v := range spec.Coverages {
				t++
				if v > 0 {
					c++
				}
			}
			total += t
			covered += c
			coverages = append(coverages, []string{fmt.Sprintf("  %s", spec.Key), fmt.Sprintf("%.1f%%", float64(c)/float64(t)*100)})
			if flgs.Long {
				keys := lo.Keys(spec.Coverages)
				sort.SliceStable(keys, func(i, j int) bool {
					if !strings.Contains(keys[i], " ") || !strings.Contains(keys[j], " ") {
						// Sort by method ( protocol buffers )
						return keys[i] < keys[j]
					}
					// Sort by path ( OpenAPI )
					mpi := strings.SplitN(keys[i], " ", 2)
					mpj := strings.SplitN(keys[j], " ", 2)
					if mpi[1] == mpj[1] {
						// Sort by method ( OpenAPI )
						return slices.Index(sortByMethod, mpi[0]) < slices.Index(sortByMethod, mpj[0])
					}
					return mpi[1] < mpj[1]
				})
				for _, k := range keys {
					v := spec.Coverages[k]
					if v == 0 {
						coverages = append(coverages, []string{color.RedString("    %s", k), ""})
						continue
					}
					coverages = append(coverages, []string{color.GreenString("    %s", k), color.HiGreenString("%d", v)})
				}
			}
		}
		if flgs.Debug {
			cmd.Println()
		}
		if len(coverages) == 0 {
			return errors.New("could not find any specs")
		}
		if err := table.Append([]string{"Total", fmt.Sprintf("%.1f%%", float64(covered)/float64(total)*100)}); err != nil {
			return err
		}
		for _, v := range coverages {
			if err := table.Append(v); err != nil {
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
	rootCmd.AddCommand(coverageCmd)
	coverageCmd.Flags().BoolVarP(&flgs.Long, "long", "l", false, flgs.Usage("Long"))
	coverageCmd.Flags().BoolVarP(&flgs.Debug, "debug", "", false, flgs.Usage("Debug"))
	coverageCmd.Flags().StringSliceVarP(&flgs.Vars, "var", "", []string{}, flgs.Usage("Vars"))
	coverageCmd.Flags().StringSliceVarP(&flgs.Runners, "runner", "", []string{}, flgs.Usage("Runners"))
	coverageCmd.Flags().StringSliceVarP(&flgs.Overlays, "overlay", "", []string{}, flgs.Usage("Overlays"))
	coverageCmd.Flags().StringSliceVarP(&flgs.Underlays, "underlay", "", []string{}, flgs.Usage("Underlays"))
	coverageCmd.Flags().StringVarP(&flgs.RunMatch, "run", "", "", flgs.Usage("RunMatch"))
	coverageCmd.Flags().StringSliceVarP(&flgs.RunIDs, "id", "", []string{}, flgs.Usage("RunIDs"))
	coverageCmd.Flags().StringSliceVarP(&flgs.RunLabels, "label", "", []string{}, flgs.Usage("RunLabels"))
	coverageCmd.Flags().BoolVarP(&flgs.SkipIncluded, "skip-included", "", false, flgs.Usage("SkipIncluded"))
	coverageCmd.Flags().StringSliceVarP(&flgs.HTTPOpenApi3s, "http-openapi3", "", []string{}, flgs.Usage("HTTPOpenApi3s"))
	coverageCmd.Flags().BoolVarP(&flgs.GRPCNoTLS, "grpc-no-tls", "", false, flgs.Usage("GRPCNoTLS"))
	coverageCmd.Flags().StringSliceVarP(&flgs.GRPCProtos, "grpc-proto", "", []string{}, flgs.Usage("GRPCProtos"))
	coverageCmd.Flags().StringSliceVarP(&flgs.GRPCImportPaths, "grpc-import-path", "", []string{}, flgs.Usage("GRPCImportPaths"))
	coverageCmd.Flags().StringSliceVarP(&flgs.GRPCBufDirs, "grpc-buf-dir", "", []string{}, flgs.Usage("GRPCBufDirs"))
	coverageCmd.Flags().StringSliceVarP(&flgs.GRPCBufLocks, "grpc-buf-lock", "", []string{}, flgs.Usage("GRPCBufLocks"))
	coverageCmd.Flags().StringSliceVarP(&flgs.GRPCBufConfigs, "grpc-buf-config", "", []string{}, flgs.Usage("GRPCBufConfigs"))
	coverageCmd.Flags().StringSliceVarP(&flgs.GRPCBufModules, "grpc-buf-module", "", []string{}, flgs.Usage("GRPCBufModules"))
	coverageCmd.Flags().StringVarP(&flgs.CacheDir, "cache-dir", "", "", flgs.Usage("CacheDir"))
	coverageCmd.Flags().StringVarP(&flgs.Format, "format", "", "", flgs.Usage("Format"))
	coverageCmd.Flags().BoolVarP(&flgs.RetainCacheDir, "retain-cache-dir", "", false, flgs.Usage("RetainCacheDir"))
	coverageCmd.Flags().StringVarP(&flgs.EnvFile, "env-file", "", "", flgs.Usage("EnvFile"))
	if err := coverageCmd.MarkFlagFilename("env-file"); err != nil {
		panic(err)
	}
}
