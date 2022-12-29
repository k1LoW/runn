package flags

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/capture"
	"github.com/spf13/cast"
)

var intRe = regexp.MustCompile(`^\-?[0-9]+$`)
var floatRe = regexp.MustCompile(`^\-?[0-9.]+$`)

type Flags struct {
	Debug           bool     `usage:"debug"`
	FailFast        bool     `usage:"fail fast"`
	SkipTest        bool     `usage:"skip \"test:\" section"`
	SkipIncluded    bool     `usage:"skip running the included runbook by itself"`
	GRPCNoTLS       bool     `usage:"disable TLS use in all gRPC runners"`
	CaptureDir      string   `usage:"destination of runbook run capture results"`
	Vars            []string `usage:"set var to runbook (\"key:value\")"`
	Runners         []string `usage:"set runner to runbook (\"key:dsn\")"`
	Overlays        []string `usage:"overlay values on the runbook"`
	Underlays       []string `usage:"lay values under the runbook"`
	Sample          int      `usage:"sample the specified number of runbooks"`
	Shuffle         string   `usage:"randomize the order of running runbooks (\"on\",\"off\",N)"`
	Parallel        string   `usage:"parallelize runs of runbooks (\"on\",\"off\",N)"`
	Random          int      `usage:"run the specified number of runbooks at random"`
	Desc            string   `usage:"description of runbook"`
	Out             string   `usage:"target path of runbook"`
	Format          string   `usage:"format of result output"`
	AndRun          bool     `usage:"run created runbook and capture the response for test"`
	LoadTConcurrent int      `usage:"number of parallel load test runs"`
	LoadTDuration   string   `usage:"load test running duration"`
	LoadTWarmUp     string   `usage:"warn-up time for load test"`
	Profile         bool     `usage:"profile runs of runbooks"`
	ProfileOut      string   `usage:"profile output path"`
	ProfileDepth    int      `usage:"depth of profile"`
	ProfileUnit     string   `usage:"-"`
	ProfileSort     string   `usage:"-"`
	CacheDir        string   `usage:"specify cache directory for remote runbooks"`
	RetainCacheDir  bool     `usage:"retain cache directory for remote runbooks"`
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

func (f *Flags) Usage(name string) string {
	field, ok := reflect.TypeOf(f).Elem().FieldByName(name)
	if !ok {
		panic(fmt.Sprintf("invalid name: %s", name))
	}
	return field.Tag.Get("usage")
}
