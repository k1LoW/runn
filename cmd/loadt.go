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

// loadtCmd represents the loadt command
var loadtCmd = &cobra.Command{
	Use:     "loadt [PATH_PATTERN]",
	Short:   "run load test using runbooks",
	Long:    `run load test using runbooks.`,
	Aliases: []string{"loadtest"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		pathp := strings.Join(args, string(filepath.ListSeparator))
		opts, err := flags.ToOpts()
		if err != nil {
			return err
		}
		o, err := runn.Load(pathp, opts...)
		if err != nil {
			return err
		}
		defer func() {
			_ = runn.RemoveCacheDir()
		}()
		d, err := duration.Parse(flags.LoadTDuration)
		if err != nil {
			return err
		}
		w, err := duration.Parse(flags.LoadTWarmUp)
		if err != nil {
			return err
		}
		s, err := setting.New(flags.LoadTConcurrent, d, w)
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

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loadtCmd)
	loadtCmd.Flags().BoolVarP(&flags.Debug, "debug", "", false, flags.Usage("Debug"))
	loadtCmd.Flags().BoolVarP(&flags.FailFast, "fail-fast", "", false, flags.Usage("FailFast"))
	loadtCmd.Flags().BoolVarP(&flags.SkipTest, "skip-test", "", false, flags.Usage("SkipTest"))
	loadtCmd.Flags().BoolVarP(&flags.SkipIncluded, "skip-included", "", false, flags.Usage("SkipIncluded"))
	loadtCmd.Flags().BoolVarP(&flags.GRPCNoTLS, "grpc-no-tls", "", false, flags.Usage("GRPCNoTLS"))
	loadtCmd.Flags().StringVarP(&flags.CaptureDir, "capture", "", "", flags.Usage("CaptureDir"))
	loadtCmd.Flags().StringSliceVarP(&flags.Vars, "var", "", []string{}, flags.Usage("Vars"))
	loadtCmd.Flags().StringSliceVarP(&flags.Runners, "runner", "", []string{}, flags.Usage("Runners"))
	loadtCmd.Flags().StringSliceVarP(&flags.Overlays, "overlay", "", []string{}, flags.Usage("Overlays"))
	loadtCmd.Flags().StringSliceVarP(&flags.Underlays, "underlay", "", []string{}, flags.Usage("Underlays"))
	loadtCmd.Flags().IntVarP(&flags.Sample, "sample", "", 0, flags.Usage("Sample"))
	loadtCmd.Flags().StringVarP(&flags.Shuffle, "shuffle", "", "off", flags.Usage("Shuffle"))
	loadtCmd.Flags().StringVarP(&flags.Parallel, "parallel", "", "off", flags.Usage("Parallel"))
	loadtCmd.Flags().IntVarP(&flags.Random, "random", "", 0, flags.Usage("Random"))

	loadtCmd.Flags().IntVarP(&flags.LoadTConcurrent, "concurrent", "", 1, flags.Usage("LoadTConcurrent"))
	loadtCmd.Flags().StringVarP(&flags.LoadTDuration, "duration", "", "10sec", flags.Usage("LoadTDuration"))
	loadtCmd.Flags().StringVarP(&flags.LoadTWarmUp, "warm-up", "", "5sec", flags.Usage("LoadTWarmUp"))
}
