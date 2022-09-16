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
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [PATH_PATTERN ...]",
	Short: "run books",
	Long:  `run books.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		total := 0
		skipped := 0
		failed := 0
		books := []runn.Option{}
		for _, p := range args {
			b, err := runn.Books(p)
			if err != nil {
				return err
			}
			books = append(books, b...)
		}
		opts := []runn.Option{
			runn.Debug(debug),
			runn.SkipTest(skipTest),
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
		for _, b := range books {
			total += 1
			desc := runn.GetDesc(b)
			o, err := runn.New(append(opts, b)...)
			if err != nil {
				fmt.Printf("%s ... %v\n", desc, red(err))
				failed += 1
				if failFast {
					return err
				}
				continue
			}
			if err := o.Run(ctx); err != nil {
				fmt.Printf("%s ... %v\n", desc, red(err))
				failed += 1
			} else {
				if o.Skipped() {
					fmt.Printf("%s ... %s\n", desc, yellow("skip"))
					skipped += 1
				} else {
					fmt.Printf("%s ... %s\n", desc, green("ok"))
				}
			}
		}
		fmt.Println("")
		var ts, fs string
		if total == 1 {
			ts = fmt.Sprintf("%d scenario", total)
		} else {
			ts = fmt.Sprintf("%d scenarios", total)
		}
		ss := fmt.Sprintf("%d skipped", skipped)
		if failed == 1 {
			fs = fmt.Sprintf("%d failure", failed)
		} else {
			fs = fmt.Sprintf("%d failures", failed)
		}
		if failed > 0 {
			_, _ = fmt.Fprintf(os.Stdout, red("%s, %s, %s\n"), ts, ss, fs)
		} else {
			_, _ = fmt.Fprintf(os.Stdout, green("%s, %s, %s\n"), ts, ss, fs)
		}
		if failed > 0 {
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
}
