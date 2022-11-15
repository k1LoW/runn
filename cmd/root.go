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
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/capture"
	"github.com/k1LoW/runn/version"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var flags = &Flags{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "runn",
	Short:        "runn is a tool for running operations following a scenario",
	Long:         `runn is a tool for running operations following a scenario.`,
	Version:      version.Version,
	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var intRe = regexp.MustCompile(`^\-?[0-9]+$`)
var floatRe = regexp.MustCompile(`^\-?[0-9.]+$`)

type Flags struct {
	Debug           bool
	FailFast        bool
	SkipTest        bool
	SkipIncluded    bool
	GRPCNoTLS       bool
	CaptureDir      string
	Vars            []string
	Runners         []string
	Overlays        []string
	Underlays       []string
	Sample          int
	Shuffle         string
	Parallel        string
	Random          int
	Desc            string
	Out             string
	Format          string
	AndRun          bool
	LoadTConcurrent int
	LoadTDuration   string
	LoadTWarmUp     string
	Profile         bool
	ProfileOut      string
	ProfileDepth    int
	ProfileUnit     string
	ProfileSort     string
}

func (f *Flags) ToOpts() ([]runn.Option, error) {
	const (
		on          = "on"
		off         = "off"
		keyValueSep = ":"
		keysSep     = "."
	)
	opts := []runn.Option{
		runn.Debug(f.Debug),
		runn.SkipTest(f.SkipTest),
		runn.SkipIncluded(f.SkipIncluded),
		runn.GRPCNoTLS(f.GRPCNoTLS),
		runn.Profile(f.Profile),
	}
	if f.Sample > 0 {
		opts = append(opts, runn.RunSample(f.Sample))
	}
	if f.Shuffle != "" {
		switch {
		case f.Shuffle == on:
			opts = append(opts, runn.RunShuffle(true, time.Now().UnixNano()))
		case f.Shuffle == off:
		default:
			seed, err := strconv.ParseInt(f.Shuffle, 10, 64)
			if err != nil {
				return nil, errors.New(`should be "on", "off" or number for seed: --shuffle`)
			}
			opts = append(opts, runn.RunShuffle(true, seed))
		}
	}
	if f.Parallel != "" {
		switch {
		case f.Parallel == on:
			opts = append(opts, runn.RunParallel(true, int64(runtime.GOMAXPROCS(0))))
		case f.Parallel == off:
		default:
			max, err := strconv.ParseInt(f.Parallel, 10, 64)
			if err != nil {
				return nil, errors.New(`should be "on", "off" or number for seed: --parallel`)
			}
			opts = append(opts, runn.RunParallel(true, max))
		}
	}
	if f.Random > 0 {
		opts = append(opts, runn.RunRandom(f.Random))
	}

	for _, v := range f.Vars {
		splitted := strings.Split(v, keyValueSep)
		if len(splitted) < 2 {
			return nil, fmt.Errorf("invalid var: %s", v)
		}
		vk := strings.Split(splitted[0], keysSep)
		vv := strings.Join(splitted[1:], keyValueSep)
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
	for _, v := range f.Runners {
		splitted := strings.Split(v, keyValueSep)
		if len(splitted) < 2 {
			return nil, fmt.Errorf("invalid var: %s", v)
		}
		vk := splitted[0]
		vv := strings.Join(splitted[1:], keyValueSep)
		opts = append(opts, runn.Runner(vk, vv))
	}
	for _, o := range f.Overlays {
		opts = append(opts, runn.Overlay(o))
	}
	sort.SliceStable(f.Underlays, func(i, j int) bool {
		return i > j
	})
	for _, u := range f.Underlays {
		opts = append(opts, runn.Underlay(u))
	}
	if f.CaptureDir != "" {
		fi, err := os.Stat(f.CaptureDir)
		if err != nil {
			return nil, err
		}
		if !fi.IsDir() {
			return nil, fmt.Errorf("%s is not directory", f.CaptureDir)
		}
		opts = append(opts, runn.Capture(capture.Runbook(f.CaptureDir)))
	}
	return opts, nil
}
