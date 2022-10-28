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
	"strings"

	"github.com/fatih/color"
	"github.com/k1LoW/runn"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [PATH_PATTERN ...]",
	Short: "run scenarios of runbooks",
	Long:  `run scenarios of runbooks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		green := color.New(color.FgGreen).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		pathp := strings.Join(args, string(filepath.ListSeparator))
		opts, err := flags.ToOpts()
		if err != nil {
			return err
		}
		opts = append(opts, runn.Capture(runn.NewCmdOut(os.Stdout)))

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
		if r.Total.Load() == 1 {
			ts = fmt.Sprintf("%d scenario", r.Total.Load())
		} else {
			ts = fmt.Sprintf("%d scenarios", r.Total.Load())
		}
		ss := fmt.Sprintf("%d skipped", r.Skipped.Load())
		if r.Failed.Load() == 1 {
			fs = fmt.Sprintf("%d failure", r.Failed.Load())
		} else {
			fs = fmt.Sprintf("%d failures", r.Failed.Load())
		}
		if r.Failed.Load() > 0 {
			_, _ = fmt.Fprintf(os.Stdout, red("%s, %s, %s\n"), ts, ss, fs)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, green("%s, %s, %s\n"), ts, ss, fs)
		}
		if r.Failed.Load() > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&flags.Debug, "debug", "", false, "debug")
	runCmd.Flags().BoolVarP(&flags.FailFast, "fail-fast", "", false, "fail fast")
	runCmd.Flags().BoolVarP(&flags.SkipTest, "skip-test", "", false, `skip "test:" section`)
	runCmd.Flags().BoolVarP(&flags.SkipIncluded, "skip-included", "", false, `skip running the included step by itself`)
	runCmd.Flags().BoolVarP(&flags.GRPCNoTLS, "grpc-no-tls", "", false, "disable TLS use in all gRPC runners")
	runCmd.Flags().StringVarP(&flags.CaptureDir, "capture", "", "", "destination of runbook run capture results")
	runCmd.Flags().StringSliceVarP(&flags.Vars, "var", "", []string{}, `set var to runbook ("key:value")`)
	runCmd.Flags().StringSliceVarP(&flags.Runners, "runner", "", []string{}, `set runner to runbook ("key:dsn")`)
	runCmd.Flags().StringSliceVarP(&flags.Overlays, "overlay", "", []string{}, "overlay values on the runbook")
	runCmd.Flags().StringSliceVarP(&flags.Underlays, "underlay", "", []string{}, "lay values under the runbook")
	runCmd.Flags().IntVarP(&flags.Sample, "sample", "", 0, "run the specified number of runbooks at random")
	runCmd.Flags().StringVarP(&flags.Shuffle, "shuffle", "", "off", `randomize the order of running runbooks ("on","off",N)`)
	runCmd.Flags().StringVarP(&flags.Parallel, "parallel", "", "off", `parallelize runs of runbooks ("on","off",N)`)
}
