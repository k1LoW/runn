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
	"path/filepath"
	"strings"

	"github.com/k1LoW/duration"
	"github.com/k1LoW/runn"
	"github.com/ryo-yamaoka/otchkiss"
	"github.com/ryo-yamaoka/otchkiss/setting"
	"github.com/spf13/cobra"
)

const reportTemplate = `
Warm up time (--warm-up)......: {{.WarmUpTime}}
Duration (--duration).........: {{.Duration}}
Concurrent (--concurrent).....: {{.MaxConcurrent}}

Total.........................: {{.TotalRequests}}
Succeeded.....................: {{.Succeeded}}
Failed........................: {{.Failed}}
Error rate....................: {{.ErrorRate}}%
RunN per seconds..............: {{.RPS}}
Latency ......................: max={{.MaxLatency}}ms min={{.MinLatency}}ms avg={{.AvgLatency}}ms med={{.MedLatency}}ms p(90)={{.Latency90p}}ms p(99)={{.Latency99p}}ms
`

// loadtCmd represents the loadt command.
var loadtCmd = &cobra.Command{
	Use:     "loadt [PATH_PATTERN]",
	Short:   "run load test using runbooks",
	Long:    `run load test using runbooks.`,
	Aliases: []string{"loadtest"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		pathp := strings.Join(args, string(filepath.ListSeparator))
		opts, err := flgs.ToOpts()
		if err != nil {
			return err
		}

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
		d, err := duration.Parse(flgs.LoadTDuration)
		if err != nil {
			return err
		}
		w, err := duration.Parse(flgs.LoadTWarmUp)
		if err != nil {
			return err
		}
		s, err := setting.New(flgs.LoadTConcurrent, d, w)
		if err != nil {
			return err
		}
		selected, err := o.SelectedOperators()
		if err != nil {
			return err
		}
		tmpl := fmt.Sprintf("\nNumber of runbooks per RunN...: %d%s", len(selected), reportTemplate)
		ot, err := otchkiss.FromConfig(o, s, 100_000_000)
		if err != nil {
			return err
		}
		if err := ot.Start(ctx); err != nil {
			return err
		}
		rep, err := ot.TemplateReport(tmpl)
		if err != nil {
			return err
		}
		cmd.Println(rep)
		if err := runn.CheckThreshold(flgs.LoadTThreshold, d, ot.Result); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loadtCmd)
	loadtCmd.Flags().BoolVarP(&flgs.Debug, "debug", "", false, flgs.Usage("Debug"))
	loadtCmd.Flags().BoolVarP(&flgs.FailFast, "fail-fast", "", false, flgs.Usage("FailFast"))
	loadtCmd.Flags().BoolVarP(&flgs.SkipTest, "skip-test", "", false, flgs.Usage("SkipTest"))
	loadtCmd.Flags().BoolVarP(&flgs.SkipIncluded, "skip-included", "", false, flgs.Usage("SkipIncluded"))
	loadtCmd.Flags().BoolVarP(&flgs.GRPCNoTLS, "grpc-no-tls", "", false, flgs.Usage("GRPCNoTLS"))
	loadtCmd.Flags().StringVarP(&flgs.CaptureDir, "capture", "", "", flgs.Usage("CaptureDir"))
	loadtCmd.Flags().StringSliceVarP(&flgs.Vars, "var", "", []string{}, flgs.Usage("Vars"))
	loadtCmd.Flags().StringSliceVarP(&flgs.Runners, "runner", "", []string{}, flgs.Usage("Runners"))
	loadtCmd.Flags().StringSliceVarP(&flgs.Overlays, "overlay", "", []string{}, flgs.Usage("Overlays"))
	loadtCmd.Flags().StringSliceVarP(&flgs.Underlays, "underlay", "", []string{}, flgs.Usage("Underlays"))
	loadtCmd.Flags().IntVarP(&flgs.Sample, "sample", "", 0, flgs.Usage("Sample"))
	loadtCmd.Flags().StringVarP(&flgs.Shuffle, "shuffle", "", "off", flgs.Usage("Shuffle"))
	loadtCmd.Flags().StringVarP(&flgs.Parallel, "parallel", "", "off", flgs.Usage("Parallel"))
	loadtCmd.Flags().IntVarP(&flgs.Random, "random", "", 0, flgs.Usage("Random"))
	loadtCmd.Flags().StringVarP(&flgs.CacheDir, "cache-dir", "", "", flgs.Usage("CacheDir"))
	loadtCmd.Flags().BoolVarP(&flgs.RetainCacheDir, "retain-cache-dir", "", false, flgs.Usage("RetainCacheDir"))

	loadtCmd.Flags().IntVarP(&flgs.LoadTConcurrent, "concurrent", "", 1, flgs.Usage("LoadTConcurrent"))
	loadtCmd.Flags().StringVarP(&flgs.LoadTDuration, "duration", "", "10sec", flgs.Usage("LoadTDuration"))
	loadtCmd.Flags().StringVarP(&flgs.LoadTWarmUp, "warm-up", "", "5sec", flgs.Usage("LoadTWarmUp"))
	loadtCmd.Flags().StringVarP(&flgs.LoadTThreshold, "threshold", "", "", flgs.Usage("LoadTThreshold"))
}
