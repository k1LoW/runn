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
	"strings"

	"github.com/k1LoW/donegroup"
	"github.com/k1LoW/runn"
	"github.com/spf13/cobra"
)

// runCmd represents the run command.
var runCmd = &cobra.Command{
	Use:   "run [PATH_PATTERN ...]",
	Short: "run scenarios of runbooks",
	Long:  `run scenarios of runbooks.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx, cancel := donegroup.WithCancel(context.Background())
		defer func() {
			cancel()
			derr := donegroup.Wait(ctx)
			err = errors.Join(err, derr)
		}()

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
		if err := o.RunN(ctx); err != nil {
			return err
		}
		r := o.Result()
		switch flgs.Format {
		case "json":
			if err := r.OutJSON(os.Stdout); err != nil {
				return err
			}
		case "none":
		default:
			// If --verbose == true, leave it to cmdout to display results
			if err := r.Out(os.Stdout, !flgs.Verbose); err != nil {
				return err
			}
		}

		if flgs.Profile {
			p, err := os.Create(filepath.Clean(flgs.ProfileOut))
			if err != nil {
				return err
			}
			defer func() {
				if err := p.Close(); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			}()
			if err := o.DumpProfile(p); err != nil {
				return err
			}
		}

		if r.HasFailure() {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&flgs.Debug, "debug", "", false, flgs.Usage("Debug"))
	runCmd.Flags().BoolVarP(&flgs.FailFast, "fail-fast", "", false, flgs.Usage("FailFast"))
	runCmd.Flags().BoolVarP(&flgs.SkipTest, "skip-test", "", false, flgs.Usage("SkipTest"))
	runCmd.Flags().BoolVarP(&flgs.SkipIncluded, "skip-included", "", false, flgs.Usage("SkipIncluded"))
	runCmd.Flags().StringSliceVarP(&flgs.HostRules, "host-rules", "", []string{}, flgs.Usage("HostRules"))
	runCmd.Flags().StringSliceVarP(&flgs.HTTPOpenApi3s, "http-openapi3", "", []string{}, flgs.Usage("HTTPOpenApi3s"))
	runCmd.Flags().BoolVarP(&flgs.GRPCNoTLS, "grpc-no-tls", "", false, flgs.Usage("GRPCNoTLS"))
	runCmd.Flags().StringSliceVarP(&flgs.GRPCProtos, "grpc-proto", "", []string{}, flgs.Usage("GRPCProtos"))
	runCmd.Flags().StringSliceVarP(&flgs.GRPCImportPaths, "grpc-import-path", "", []string{}, flgs.Usage("GRPCImportPaths"))
	runCmd.Flags().StringSliceVarP(&flgs.GRPCBufDirs, "grpc-buf-dir", "", []string{}, flgs.Usage("GRPCBufDirs"))
	runCmd.Flags().StringSliceVarP(&flgs.GRPCBufLocks, "grpc-buf-lock", "", []string{}, flgs.Usage("GRPCBufLocks"))
	runCmd.Flags().StringSliceVarP(&flgs.GRPCBufConfigs, "grpc-buf-config", "", []string{}, flgs.Usage("GRPCBufConfigs"))
	runCmd.Flags().StringSliceVarP(&flgs.GRPCBufModules, "grpc-buf-module", "", []string{}, flgs.Usage("GRPCBufModules"))
	runCmd.Flags().StringVarP(&flgs.CaptureDir, "capture", "", "", flgs.Usage("CaptureDir"))
	runCmd.Flags().StringSliceVarP(&flgs.Vars, "var", "", []string{}, flgs.Usage("Vars"))
	runCmd.Flags().StringSliceVarP(&flgs.Runners, "runner", "", []string{}, flgs.Usage("Runners"))
	runCmd.Flags().StringSliceVarP(&flgs.Overlays, "overlay", "", []string{}, flgs.Usage("Overlays"))
	runCmd.Flags().StringSliceVarP(&flgs.Underlays, "underlay", "", []string{}, flgs.Usage("Underlays"))
	runCmd.Flags().StringVarP(&flgs.RunMatch, "run", "", "", flgs.Usage("RunMatch"))
	runCmd.Flags().StringSliceVarP(&flgs.RunIDs, "id", "", []string{}, flgs.Usage("RunIDs"))
	runCmd.Flags().StringSliceVarP(&flgs.RunLabels, "label", "", []string{}, flgs.Usage("RunLabels"))
	runCmd.Flags().IntVarP(&flgs.Sample, "sample", "", 0, flgs.Usage("Sample"))
	runCmd.Flags().StringVarP(&flgs.Shuffle, "shuffle", "", "off", flgs.Usage("Shuffle"))
	runCmd.Flags().StringVarP(&flgs.Concurrent, "concurrent", "", "off", flgs.Usage("Concurrent"))
	runCmd.Flags().IntVarP(&flgs.ShardIndex, "shard-index", "", 0, flgs.Usage("ShardIndex"))
	runCmd.Flags().IntVarP(&flgs.ShardN, "shard-n", "", 0, flgs.Usage("ShardN"))
	runCmd.Flags().IntVarP(&flgs.Random, "random", "", 0, flgs.Usage("Random"))
	runCmd.Flags().StringVarP(&flgs.Format, "format", "", "", flgs.Usage("Format"))
	runCmd.Flags().BoolVarP(&flgs.Profile, "profile", "", false, flgs.Usage("Profile"))
	runCmd.Flags().StringVarP(&flgs.ProfileOut, "profile-out", "", "runn.prof", flgs.Usage("ProfileOut"))
	runCmd.Flags().StringVarP(&flgs.CacheDir, "cache-dir", "", "", flgs.Usage("CacheDir"))
	runCmd.Flags().BoolVarP(&flgs.RetainCacheDir, "retain-cache-dir", "", false, flgs.Usage("RetainCacheDir"))
	runCmd.Flags().StringVarP(&flgs.WaitTimeout, "wait-timeout", "", "10sec", flgs.Usage("WaitTimeout"))
	runCmd.Flags().BoolVarP(&flgs.Verbose, "verbose", "", false, flgs.Usage("Verbose"))
	runCmd.Flags().BoolVarP(&flgs.Attach, "attach", "", false, flgs.Usage("Attach"))
}
