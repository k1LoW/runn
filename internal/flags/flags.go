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

	"github.com/k1LoW/duration"
	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/capture"
	"github.com/k1LoW/runn/internal/store"
	"github.com/spf13/cast"
)

var intRe = regexp.MustCompile(`^\-?[0-9]+$`)
var floatRe = regexp.MustCompile(`^\-?[0-9.]+$`)

type Flags struct {
	Debug           bool     `usage:"debug"`
	Long            bool     `usage:"long format"`
	FailFast        bool     `usage:"fail fast"`
	SkipTest        bool     `usage:"skip \"test:\" section"`
	SkipIncluded    bool     `usage:"skip running the included runbook by itself"`
	RunMatch        string   `usage:"run all runbooks with a matching file path, treating the value passed to the option as an unanchored regular expression"`
	RunIDs          []string `usage:"run the matching runbooks in order if there is only one runbook with a forward matching ID"`
	RunLabels       []string `usage:"run all runbooks matching the label specification"`
	HTTPOpenApi3s   []string `usage:"set the path to the OpenAPI v3 document for HTTP runners (\"path/to/spec.yml\" or \"key:path/to/spec.yml\")"`
	GRPCNoTLS       bool     `usage:"disable TLS use in all gRPC runners"`
	GRPCProtos      []string `usage:"set the name of proto source for gRPC runners"`
	GRPCImportPaths []string `usage:"set the path to the directory where proto sources can be imported for gRPC runners"`
	GRPCBufDirs     []string `usage:"set the path to the buf directory for gRPC runners"`
	GRPCBufLocks    []string `usage:"set the path to buf.lock for gRPC runners"`
	GRPCBufConfigs  []string `usage:"set the path to buf.yaml for gRPC runners"`
	GRPCBufModules  []string `usage:"set the buf modules for gRPC runners (\"buf.build/owner/repository\" or \"buf.build/owner/repository/tree/branch-or-commit\")"`
	CaptureDir      string   `usage:"destination of runbook run capture results"`
	Vars            []string `usage:"set var to runbook (\"key:value\")"`
	Runners         []string `usage:"set runner to runbook (\"key:dsn\")"`
	Overlays        []string `usage:"overlay values on the runbook"`
	Underlays       []string `usage:"lay values under the runbook"`
	Sample          int      `usage:"sample the specified number of runbooks"`
	Shuffle         string   `usage:"randomize the order of running runbooks (\"on\",\"off\",N)"`
	Concurrent      string   `usage:"run runbooks concurrently (\"on\",\"off\",N)"`
	ShardIndex      int      `usage:"index of distributed runbooks"`
	ShardN          int      `usage:"number of shards for distributing runbooks"`
	Random          int      `usage:"run the specified number of runbooks at random"`
	Desc            string   `usage:"description of runbook"`
	Out             string   `usage:"target path of runbook"`
	Format          string   `usage:"format of result output"`
	AndRun          bool     `usage:"run created runbook and capture the response for test"`
	LoadTConcurrent int      `usage:"number of concurrent load test runs. 0 means unlimited"`
	LoadTDuration   string   `usage:"load test running duration"`
	LoadTWarmUp     string   `usage:"warn-up time for load test"`
	LoadTThreshold  string   `usage:"if this threshold condition is not met, loadt command returns exit status 1 (EXIT_FAILURE)"`
	LoadTMaxRPS     int      `usage:"max RunN per second for load test. 0 means unlimited"`
	Profile         bool     `usage:"profile runs of runbooks"`
	ProfileOut      string   `usage:"profile output path"`
	ProfileDepth    int      `usage:"depth of profile"`
	ProfileUnit     string   `usage:"-"`
	ProfileSort     string   `usage:"-"`
	Attach          bool     `usage:"attach to runn process"`
	CacheDir        string   `usage:"specify cache directory for remote runbooks"`
	RetainCacheDir  bool     `usage:"retain cache directory for remote runbooks"`
	Scopes          []string `usage:"additional scopes for runn"`
	HostRules       []string `usage:"host rules for runn. (\"host rule,host rule,...\")"`
	WaitTimeout     string   `usage:"timeout for waiting for cleanup process after running runbooks"`
	EnvFile         string   `usage:"load environment variables from a file"`
	ForceColor      bool     `usage:"force colorized output even in non-tty output streams"`
	Verbose         bool     `usage:"verbose"`
	Coverage        bool     `usage:"coverage for OpenAPI spec and protocol buffers"`
	CoverageOut     string   `usage:"coverage output path (JSON format)"`
}

func (f *Flags) ToOpts() ([]runn.Option, error) {
	if err := runn.LoadEnvFile(f.EnvFile); err != nil {
		return nil, err
	}
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
		runn.HTTPOpenApi3s(f.HTTPOpenApi3s),
		runn.GRPCNoTLS(f.GRPCNoTLS),
		runn.GRPCProtos(f.GRPCProtos),
		runn.GRPCImportPaths(f.GRPCImportPaths),
		runn.GRPCBufDir(f.GRPCBufDirs...),
		runn.GRPCBufLock(f.GRPCBufLocks...),
		runn.GRPCBufConfig(f.GRPCBufConfigs...),
		runn.GRPCBufModule(f.GRPCBufModules...),
		runn.Profile(f.Profile),
		runn.Scopes(f.Scopes...),
		runn.HostRules(f.HostRules...),
		runn.RunLabel(f.RunLabels...),
		runn.FailFast(f.FailFast),
		runn.Attach(f.Attach),
	}

	// STDIN
	if err := store.SetStdin(os.Stdin); err != nil {
		return nil, err
	}

	// From flags
	opts = append(opts, runn.RunID(f.RunIDs...))

	if f.WaitTimeout != "" {
		wt, err := duration.Parse(f.WaitTimeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, runn.WaitTimeout(wt))
	}

	if f.RunMatch != "" {
		opts = append(opts, runn.RunMatch(f.RunMatch))
	}
	if f.Sample > 0 {
		opts = append(opts, runn.RunSample(f.Sample))
	}
	if f.Shuffle != "" {
		switch f.Shuffle {
		case on:
			opts = append(opts, runn.RunShuffle(true, time.Now().UnixNano()))
		case off:
		default:
			seed, err := strconv.ParseInt(f.Shuffle, 10, 64)
			if err != nil {
				return nil, errors.New(`should be "on", "off" or number for seed: --shuffle`)
			}
			opts = append(opts, runn.RunShuffle(true, seed))
		}
	}
	if f.Concurrent != "" {
		if f.Attach && f.Concurrent != off {
			return nil, errors.New("cannot use --concurrent with --attach")
		}
		switch f.Concurrent {
		case on:
			opts = append(opts, runn.RunConcurrent(true, runtime.GOMAXPROCS(0)))
		case off:
		default:
			max, err := strconv.Atoi(f.Concurrent)
			if err != nil {
				return nil, errors.New(`should be "on", "off" or number for concurrent: --concurrent`)
			}
			opts = append(opts, runn.RunConcurrent(true, max))
		}
	}
	if f.Random > 0 {
		opts = append(opts, runn.RunRandom(f.Random))
	}
	if f.ShardN > 0 {
		opts = append(opts, runn.RunShard(f.ShardN, f.ShardIndex))
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
	if f.Format == "" {
		opts = append(opts, runn.Capture(runn.NewCmdOut(os.Stdout, f.Verbose)))
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
