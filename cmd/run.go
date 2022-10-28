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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/capture"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
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
		opts, err := collectOpts()
		if err != nil {
			return err
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
	runCmd.Flags().BoolVarP(&debug, "debug", "", false, "debug")
	runCmd.Flags().BoolVarP(&failFast, "fail-fast", "", false, "fail fast")
	runCmd.Flags().BoolVarP(&skipTest, "skip-test", "", false, `skip "test:" section`)
	runCmd.Flags().BoolVarP(&skipIncluded, "skip-included", "", false, `skip running the included step by itself`)
	runCmd.Flags().BoolVarP(&grpcNoTLS, "grpc-no-tls", "", false, "disable TLS use in all gRPC runners")
	runCmd.Flags().StringVarP(&captureDir, "capture", "", "", "destination of runbook run capture results")
	runCmd.Flags().StringSliceVarP(&vars, "var", "", []string{}, `set var to runbook ("key:value")`)
	runCmd.Flags().StringSliceVarP(&overlays, "overlay", "", []string{}, "overlay values on the runbook")
	runCmd.Flags().StringSliceVarP(&underlays, "underlay", "", []string{}, "lay values under the runbook")
	runCmd.Flags().IntVarP(&sample, "sample", "", 0, "run the specified number of runbooks at random")
	runCmd.Flags().StringVarP(&shuffle, "shuffle", "", "off", `randomize the order of running runbooks ("on","off",N)`)
	runCmd.Flags().StringVarP(&parallel, "parallel", "", "off", `parallelize runs of runbooks ("on","off",N)`)
}

var intRe = regexp.MustCompile(`^\-?[0-9]+$`)
var floatRe = regexp.MustCompile(`^\-?[0-9.]+$`)

func collectOpts() ([]runn.Option, error) {
	opts := []runn.Option{
		runn.Debug(debug),
		runn.SkipTest(skipTest),
		runn.SkipIncluded(skipIncluded),
		runn.GRPCNoTLS(grpcNoTLS),
		runn.Capture(runn.NewCmdOut(os.Stdout)),
	}
	if sample > 0 {
		opts = append(opts, runn.RunSample(sample))
	}
	if shuffle != "" {
		switch {
		case shuffle == "on":
			opts = append(opts, runn.RunShuffle(true, time.Now().UnixNano()))
		case shuffle == "off":
		default:
			seed, err := strconv.ParseInt(shuffle, 10, 64)
			if err != nil {
				return nil, errors.New(`should be "on", "off" or number for seed: --shuffle`)
			}
			opts = append(opts, runn.RunShuffle(true, seed))
		}
	}
	if parallel != "" {
		switch {
		case parallel == "on":
			opts = append(opts, runn.RunParallel(true, int64(runtime.GOMAXPROCS(0))))
		case parallel == "off":
		default:
			max, err := strconv.ParseInt(parallel, 10, 64)
			if err != nil {
				return nil, errors.New(`should be "on", "off" or number for seed: --parallel`)
			}
			opts = append(opts, runn.RunParallel(true, max))
		}
	}
	for _, v := range vars {
		splitted := strings.Split(v, ":")
		if len(splitted) < 2 {
			return nil, fmt.Errorf("invalid var: %s", v)
		}
		vk := strings.Split(splitted[0], ".")
		vv := strings.Join(splitted[1:], ":")
		switch {
		case intRe.MatchString(vv):
			vvv, err := cast.ToIntE(vv)
			if err == nil {
				opts = append(opts, runn.Var(vk, vvv))
				continue
			}
		case floatRe.MatchString(vv):
			vvv, err := cast.ToFloat64E(vv)
			if err == nil {
				opts = append(opts, runn.Var(vk, vvv))
				continue
			}
		}
		opts = append(opts, runn.Var(vk, vv))
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
			return nil, err
		}
		if !fi.IsDir() {
			return nil, fmt.Errorf("%s is not directory", captureDir)
		}
		opts = append(opts, runn.Capture(capture.Runbook(captureDir)))
	}
	return opts, nil
}
