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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/capture"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	failFast   bool
	skipTest   bool
	captureDir string
	overlays   []string
	underlays  []string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [PATH_PATTERN ...]",
	Short: "run books",
	Long:  `run books.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		pathp := strings.Join(args, string(filepath.ListSeparator))
		opts := []runn.Option{
			runn.Debug(debug),
			runn.SkipTest(skipTest),
			runn.Capture(runn.NewCmdOut(os.Stdout)),
		}
		for _, o := range overlays {
			opts = append(opts, runn.Overlay(o))
		}
		sort.SliceStable(underlays, func(i, j int) bool {
			return i > j
		})
		for _, u := range underlays {
			opts = append(opts, runn.Underlay(u))
		}
		if captureDir != "" {
			fi, err := os.Stat(captureDir)
			if err != nil {
				return err
			}
			if !fi.IsDir() {
				return fmt.Errorf("%s is not directory", captureDir)
			}
			opts = append(opts, runn.Capture(capture.Runbook(captureDir)))
		}
		o, err := runn.Load(pathp, opts...)
		if err != nil {
			return err
		}
		if err := o.RunN(ctx); err != nil {
			return err
		}
		fmt.Println("")
		r := o.Result()
		var ts, fs string
		if r.Total == 1 {
			ts = fmt.Sprintf("%d scenario", r.Total)
		} else {
			ts = fmt.Sprintf("%d scenarios", r.Total)
		}
		ss := fmt.Sprintf("%d skipped", r.Skipped)
		if r.Failed == 1 {
			fs = fmt.Sprintf("%d failure", r.Failed)
		} else {
			fs = fmt.Sprintf("%d failures", r.Failed)
		}
		if r.Failed > 0 {
			_, _ = fmt.Fprintf(os.Stdout, red("%s, %s, %s\n"), ts, ss, fs)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, green("%s, %s, %s\n"), ts, ss, fs)
		}
		if r.Failed > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&debug, "debug", "", false, "debug")
	runCmd.Flags().BoolVarP(&failFast, "fail-fast", "", false, "fail fast")
	runCmd.Flags().BoolVarP(&skipTest, "skip-test", "", false, "skip test")
	runCmd.Flags().StringVarP(&captureDir, "capture", "", "", "destination of runbook run capture results")
	runCmd.Flags().StringSliceVarP(&overlays, "overlay", "", []string{}, "overlay values on the runbook")
	runCmd.Flags().StringSliceVarP(&underlays, "underlay", "", []string{}, "lay values under the runbook")
}
