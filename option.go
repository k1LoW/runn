package runn

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Songmu/prompter"
	"github.com/k1LoW/duration"
	"github.com/k1LoW/runn/builtin"
	"github.com/k1LoW/sshc/v4"
	"github.com/spf13/cast"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Option func(*book) error

var ErrNilBook = errors.New("runbook is nil")

// Book - Load runbook.
func Book(path string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		loaded, err := loadBook(path, nil)
		if err != nil {
			return err
		}
		return bk.merge(loaded)
	}
}

// Overlay - Overlay values on a runbook.
func Overlay(path string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if len(bk.rawSteps) == 0 {
			return errors.New("overlays are unusable without its base runbook")
		}
		loaded, err := loadBook(path, nil)
		if err != nil {
			return err
		}
		bk.desc = loaded.desc
		bk.ifCond = loaded.ifCond
		if len(loaded.rawSteps) > 0 {
			if bk.useMap != loaded.useMap {
				return errors.New("only runbooks of the same type can be layered")
			}
		}
		for k, r := range loaded.runners {
			bk.runners[k] = r
		}
		for k, r := range loaded.httpRunners {
			bk.httpRunners[k] = r
		}
		for k, r := range loaded.dbRunners {
			bk.dbRunners[k] = r
		}
		for k, r := range loaded.grpcRunners {
			bk.grpcRunners[k] = r
		}
		for k, r := range loaded.cdpRunners {
			bk.cdpRunners[k] = r
		}
		for k, r := range loaded.sshRunners {
			bk.sshRunners[k] = r
		}
		for k, v := range loaded.vars {
			bk.vars[k] = v
		}
		for k, e := range loaded.runnerErrs {
			bk.runnerErrs[k] = e
		}
		bk.rawSteps = append(bk.rawSteps, loaded.rawSteps...)
		bk.stepKeys = append(bk.stepKeys, loaded.stepKeys...)
		bk.debug = loaded.debug
		bk.skipTest = loaded.skipTest
		bk.loop = loaded.loop
		bk.concurrency = loaded.concurrency
		bk.grpcNoTLS = loaded.grpcNoTLS
		bk.interval = loaded.interval
		return nil
	}
}

// Underlay - Lay values under the runbook.
func Underlay(path string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if len(bk.rawSteps) == 0 {
			return errors.New("underlays are unusable without its base runbook")
		}
		loaded, err := loadBook(path, nil)
		if err != nil {
			return err
		}
		if bk.desc == "" {
			bk.desc = loaded.desc
		}
		if bk.ifCond == "" {
			bk.ifCond = loaded.ifCond
		}
		if len(loaded.rawSteps) > 0 {
			if bk.useMap != loaded.useMap {
				return errors.New("only runbooks of the same type can be layered")
			}
		}
		for k, r := range loaded.runners {
			if _, ok := bk.runners[k]; !ok {
				bk.runners[k] = r
			}
		}
		for k, r := range loaded.httpRunners {
			if _, ok := bk.httpRunners[k]; !ok {
				bk.httpRunners[k] = r
			}
		}
		for k, r := range loaded.dbRunners {
			if _, ok := bk.dbRunners[k]; !ok {
				bk.dbRunners[k] = r
			}
		}
		for k, r := range loaded.grpcRunners {
			if _, ok := bk.grpcRunners[k]; !ok {
				bk.grpcRunners[k] = r
			}
		}
		for k, r := range loaded.cdpRunners {
			if _, ok := bk.cdpRunners[k]; !ok {
				bk.cdpRunners[k] = r
			}
		}
		for k, r := range loaded.sshRunners {
			if _, ok := bk.sshRunners[k]; !ok {
				bk.sshRunners[k] = r
			}
		}
		for k, v := range loaded.vars {
			if _, ok := bk.vars[k]; !ok {
				bk.vars[k] = v
			}
		}
		for k, e := range loaded.runnerErrs {
			bk.runnerErrs[k] = e
		}
		bk.rawSteps = append(loaded.rawSteps, bk.rawSteps...)
		bk.stepKeys = append(loaded.stepKeys, bk.stepKeys...)
		if bk.intervalStr == "" {
			bk.interval = loaded.interval
		}
		bk.stdout = loaded.stdout
		bk.stderr = loaded.stderr
		return nil
	}
}

// Desc - Set description to runbook.
func Desc(desc string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.desc = desc
		return nil
	}
}

// Runner - Set runner to runbook.
func Runner(name, dsn string, opts ...httpRunnerOption) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		delete(bk.runnerErrs, name)
		if len(opts) == 0 {
			if err := validateRunnerKey(name); err != nil {
				return err
			}
			if err := bk.parseRunner(name, dsn); err != nil {
				bk.runnerErrs[name] = err
			}
			return nil
		}
		// HTTP Runner
		c := &httpRunnerConfig{}
		for _, opt := range opts {
			if err := opt(c); err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
		}
		r, err := newHTTPRunner(name, dsn)
		if err != nil {
			bk.runnerErrs[name] = err
			return nil
		}
		if c.NotFollowRedirect {
			r.client.CheckRedirect = notFollowRedirectFn
		}
		r.multipartBoundary = c.MultipartBoundary
		if c.Timeout != "" {
			r.client.Timeout, err = duration.Parse(c.Timeout)
			if err != nil {
				return fmt.Errorf("timeout in HttpRunnerConfig is invalid: %w", err)
			}
		}
		if c.OpenApi3DocLocation != "" {
			v, err := newHttpValidator(c)
			if err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
			r.validator = v
		}
		bk.httpRunners[name] = r
		return nil
	}
}

// HTTPRunner - Set HTTP runner to runbook.
func HTTPRunner(name, endpoint string, client *http.Client, opts ...httpRunnerOption) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		delete(bk.runnerErrs, name)
		root, err := bk.generateOperatorRoot()
		if err != nil {
			return err
		}
		r, err := newHTTPRunner(name, endpoint)
		if err != nil {
			return err
		}
		r.client = client
		bk.httpRunners[name] = r
		if len(opts) == 0 {
			return nil
		}
		c := &httpRunnerConfig{}
		for _, opt := range opts {
			if err := opt(c); err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
		}

		if c.NotFollowRedirect {
			r.client.CheckRedirect = notFollowRedirectFn
		}
		r.multipartBoundary = c.MultipartBoundary
		if c.OpenApi3DocLocation != "" && !strings.HasPrefix(c.OpenApi3DocLocation, "https://") && !strings.HasPrefix(c.OpenApi3DocLocation, "http://") && !strings.HasPrefix(c.OpenApi3DocLocation, "/") {
			c.OpenApi3DocLocation = fp(c.OpenApi3DocLocation, root)
		}
		if c.CACert != "" {
			b, err := readFile(fp(c.CACert, root))
			if err != nil {
				return err
			}
			r.cacert = b
		}
		if c.Cert != "" {
			b, err := readFile(fp(c.Cert, root))
			if err != nil {
				return err
			}
			r.cert = b
		}
		if c.Key != "" {
			b, err := readFile(fp(c.Key, root))
			if err != nil {
				return err
			}
			r.key = b
		}
		r.skipVerify = c.SkipVerify
		if c.Timeout != "" {
			r.client.Timeout, err = duration.Parse(c.Timeout)
			if err != nil {
				return fmt.Errorf("timeout in HttpRunnerConfig is invalid: %w", err)
			}
		}
		r.useCookie = c.UseCookie
		r.trace = c.Trace

		hv, err := newHttpValidator(c)
		if err != nil {
			bk.runnerErrs[name] = err
			return nil
		}
		r.validator = hv
		return nil
	}
}

// HTTPRunnerWithHandler - Set HTTP runner to runbook with http.Handler.
func HTTPRunnerWithHandler(name string, h http.Handler, opts ...httpRunnerOption) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		delete(bk.runnerErrs, name)
		r, err := newHTTPRunnerWithHandler(name, h)
		if err != nil {
			bk.runnerErrs[name] = err
			return nil
		}
		if len(opts) > 0 {
			c := &httpRunnerConfig{}
			for _, opt := range opts {
				if err := opt(c); err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
			}
			if c.NotFollowRedirect {
				bk.runnerErrs[name] = errors.New("runn.HTTPRunnerWithHandler does not support option NotFollowRedirect")
				return nil
			}
			r.multipartBoundary = c.MultipartBoundary
			if c.Timeout != "" {
				r.client.Timeout, err = duration.Parse(c.Timeout)
				if err != nil {
					return fmt.Errorf("timeout in HttpRunnerConfig is invalid: %w", err)
				}
			}
			v, err := newHttpValidator(c)
			if err != nil {
				bk.runnerErrs[name] = err
				return nil
			}
			r.validator = v
		}
		bk.httpRunners[name] = r
		return nil
	}
}

// DBRunner - Set DB runner to runbook.
func DBRunner(name string, client Querier) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		delete(bk.runnerErrs, name)
		nt, err := nestTx(client)
		if err != nil {
			return err
		}
		bk.dbRunners[name] = &dbRunner{
			name:   name,
			client: nt,
		}
		return nil
	}
}

// DBRunnerWithOptions - Set DB runner to runbook using options.
func DBRunnerWithOptions(name, dsn string, opts ...dbRunnerOption) Option {
	return func(bk *book) error {
		delete(bk.runnerErrs, name)
		r, err := newDBRunner(name, dsn)
		if err != nil {
			return err
		}
		if len(opts) > 0 {
			c := &dbRunnerConfig{}
			for _, opt := range opts {
				if err := opt(c); err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
			}
			r.trace = c.Trace
		}
		bk.dbRunners[name] = r
		return nil
	}
}

// GrpcRunner - Set gRPC runner to runbook.
func GrpcRunner(name string, cc *grpc.ClientConn) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		delete(bk.runnerErrs, name)
		r := &grpcRunner{
			name: name,
			cc:   cc,
			mds:  map[string]protoreflect.MethodDescriptor{},
		}
		bk.grpcRunners[name] = r
		return nil
	}
}

// GrpcRunnerWithOptions - Set gRPC runner to runbook using options.
func GrpcRunnerWithOptions(name, target string, opts ...grpcRunnerOption) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		delete(bk.runnerErrs, name)
		r := &grpcRunner{
			name:   name,
			target: target,
			mds:    map[string]protoreflect.MethodDescriptor{},
		}
		if len(opts) > 0 {
			c := &grpcRunnerConfig{}
			for _, opt := range opts {
				if err := opt(c); err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
			}
			r.tls = c.TLS
			if len(c.cacert) != 0 {
				r.cacert = c.cacert
			} else if c.CACert != "" {
				b, err := readFile(c.CACert)
				if err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
				r.cacert = b
			}
			if len(c.cert) != 0 {
				r.cert = c.cert
			} else if c.Cert != "" {
				b, err := readFile(c.Cert)
				if err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
				r.cert = b
			}
			if len(c.key) != 0 {
				r.key = c.key
			} else if c.Key != "" {
				b, err := readFile(c.Key)
				if err != nil {
					bk.runnerErrs[name] = err
					return nil
				}
				r.key = b
			}
			r.importPaths = c.ImportPaths
			r.protos = c.Protos
			r.skipVerify = c.SkipVerify
		}
		bk.grpcRunners[name] = r
		return nil
	}
}

// SSHRunner - Set SSH runner to runbook.
func SSHRunner(name string, client *ssh.Client) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		delete(bk.runnerErrs, name)
		r := &sshRunner{
			name:   name,
			client: client,
		}
		bk.sshRunners[name] = r
		return nil
	}
}

// SSHRunnerWithOptions - Set SSH runner to runbook using options.
func SSHRunnerWithOptions(name string, opts ...sshRunnerOption) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		delete(bk.runnerErrs, name)
		c := &sshRunnerConfig{}
		for _, opt := range opts {
			if err := opt(c); err != nil {
				return err
			}
		}
		if err := c.validate(); err != nil {
			return fmt.Errorf("invalid SSH runner %q: %w", name, err)
		}
		host := c.Host
		if host == "" {
			host = c.Hostname
		}
		var opts []sshc.Option
		if c.SSHConfig != "" {
			p := c.SSHConfig
			if !strings.HasPrefix(c.SSHConfig, "/") {
				p = filepath.Join(filepath.Dir(bk.path), c.SSHConfig)
			}
			opts = append(opts, sshc.ClearConfig(), sshc.ConfigPath(p))
		}
		if c.Hostname != "" {
			opts = append(opts, sshc.Hostname(c.Hostname))
		}
		if c.User != "" {
			opts = append(opts, sshc.User(c.User))
		}
		if c.Port != 0 {
			opts = append(opts, sshc.Port(c.Port))
		}
		if c.IdentityFile != "" {
			p := c.IdentityFile
			if !strings.HasPrefix(c.IdentityFile, "/") {
				p = filepath.Join(filepath.Dir(bk.path), c.IdentityFile)
			}
			b, err := readFile(p)
			if err != nil {
				return err
			}
			opts = append(opts, sshc.IdentityKey(b))
		} else if c.IdentityKey != "" {
			opts = append(opts, sshc.IdentityKey([]byte(repairKey(c.IdentityKey))))
		}
		var lf *sshLocalForward
		if c.LocalForward != "" {
			c.KeepSession = true
			if strings.Count(c.LocalForward, ":") != 2 {
				return fmt.Errorf("invalid SSH runner: %q: invalid localForward option: %s", name, c.LocalForward)
			}
			splitted := strings.SplitN(c.LocalForward, ":", 2)
			lf = &sshLocalForward{
				local:  fmt.Sprintf("127.0.0.1:%s", splitted[0]),
				remote: splitted[1],
			}
		}
		opts = append(opts, sshc.AuthMethod(sshKeyboardInteractive(c.KeyboardInteractive)))

		client, err := sshc.NewClient(host, opts...)
		if err != nil {
			return err
		}

		r := &sshRunner{
			name:         name,
			client:       client,
			keepSession:  c.KeepSession,
			localForward: lf,
		}

		if r.keepSession {
			if err := r.startSession(); err != nil {
				return err
			}
		}

		bk.sshRunners[name] = r
		return nil
	}
}

// T - Acts as test helper.
func T(t *testing.T) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.t = t
		return nil
	}
}

// Var - Set variable to runner.
func Var(k any, v any) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		root, err := bk.generateOperatorRoot()
		if err != nil {
			return err
		}
		ev, err := evaluateSchema(v, root, nil)
		if err != nil {
			return err
		}
		switch kk := k.(type) {
		case string:
			bk.vars[kk] = ev
		case []string:
			vars := bk.vars
			for _, kkk := range kk[:len(kk)-1] {
				_, ok := vars[kkk]
				if !ok {
					vars[kkk] = map[string]any{}
				}
				m, ok := vars[kkk].(map[string]any)
				if !ok {
					// clear current vars to override
					vars[kkk] = map[string]any{}
					m, _ = vars[kkk].(map[string]any)
				}
				vars = m
			}
			vars[kk[len(kk)-1]] = ev
		default:
			return fmt.Errorf("invalid key of var: %v", k)
		}
		return nil
	}
}

// Func - Set function to runner.
func Func(k string, v any) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.funcs[k] = v
		return nil
	}
}

// Debug - Enable debug output.
func Debug(debug bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if !bk.debug {
			bk.debug = debug
		}
		return nil
	}
}

// Profile - Enable profile output.
func Profile(profile bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if !bk.profile {
			bk.profile = profile
		}
		return nil
	}
}

// Interval - Set interval between steps.
func Interval(d time.Duration) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if d < 0 {
			return fmt.Errorf("invalid interval: %s", d)
		}
		bk.interval = d
		return nil
	}
}

// FailFast - Enable fail-fast.
func FailFast(enable bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.failFast = enable
		return nil
	}
}

// SkipIncluded - Skip running the included step by itself.
func SkipIncluded(enable bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.skipIncluded = enable
		return nil
	}
}

// SkipTest - Skip test section.
func SkipTest(enable bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if !bk.skipTest {
			bk.skipTest = enable
		}
		return nil
	}
}

// Force - Force all steps to run.
func Force(enable bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if !bk.force {
			bk.force = enable
		}
		return nil
	}
}

// Trace - Add tokens for tracing to headers and queries by default.
func Trace(enable bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if !bk.trace {
			bk.trace = enable
		}
		return nil
	}
}

// HTTPOpenApi3 - Set the path of OpenAPI Document for HTTP runners.
// Deprecated: Use HTTPOpenApi3s instead.
func HTTPOpenApi3(l string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.openApi3DocLocations = []string{l}
		return nil
	}
}

// HTTPOpenApi3s - Set the path of OpenAPI Document for HTTP runners.
func HTTPOpenApi3s(locations []string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.openApi3DocLocations = locations
		return nil
	}
}

// GRPCNoTLS - Disable TLS use in all gRPC runners.
func GRPCNoTLS(noTLS bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.grpcNoTLS = noTLS
		return nil
	}
}

// GRPCProtos - Set the name of proto source for gRPC runners.
func GRPCProtos(protos []string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.grpcProtos = protos
		return nil
	}
}

// GRPCImportPaths - Set the path to the directory where proto sources can be imported for gRPC runners.
func GRPCImportPaths(paths []string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.grpcImportPaths = paths
		return nil
	}
}

// BeforeFunc - Register the function to be run before the runbook is run.
func BeforeFunc(fn func(*RunResult) error) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.beforeFuncs = append(bk.beforeFuncs, fn)
		return nil
	}
}

// AfterFunc - Register the function to be run after the runbook is run.
func AfterFunc(fn func(*RunResult) error) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.afterFuncs = append(bk.afterFuncs, fn)
		return nil
	}
}

// AfterFuncIf - Register the function to be run after the runbook is run if condition is true.
func AfterFuncIf(fn func(*RunResult) error, ifCond string) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.afterFuncs = append(bk.afterFuncs, func(r *RunResult) error {
			tf, err := EvalCond(ifCond, r.Store)
			if err != nil {
				return err
			}
			if !tf {
				return nil
			}
			return fn(r)
		})
		return nil
	}
}

// Capture - Register the capturer to capture steps.
func Capture(c Capturer) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.capturers = append(bk.capturers, c)
		return nil
	}
}

// RunMatch - Run only runbooks with matching paths.
func RunMatch(m string) Option { //nostyle:repetition
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		re, err := regexp.Compile(m)
		if err != nil {
			return err
		}
		bk.runMatch = re
		return nil
	}
}

// RunID - Run the matching runbook if there is only one runbook with a forward matching ID.
func RunID(id string) Option { //nostyle:repetition
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.runID = id
		return nil
	}
}

// RunSample - Sample the specified number of runbooks.
func RunSample(n int) Option { //nostyle:repetition
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if n <= 0 {
			return fmt.Errorf("sample must be greater than 0: %d", n)
		}
		bk.runSample = n
		return nil
	}
}

// RunShard - Distribute runbooks into a specified number of shards and run the specified shard of them.
func RunShard(n, i int) Option { //nostyle:repetition
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if n <= 0 {
			return fmt.Errorf("the number of divisions is greater than 0: %d", n)
		}
		if i < 0 {
			return fmt.Errorf("the index of divisions is greater than or equal to 0: %d", i)
		}
		if i >= n {
			return fmt.Errorf("the index of divisions is less than the number of distributions (%d): %d", n, i)
		}
		bk.runShardIndex = i
		bk.runShardN = n
		return nil
	}
}

// RunShuffle - Randomize the order of running runbooks.
func RunShuffle(enable bool, seed int64) Option { //nostyle:repetition
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.runShuffle = enable
		bk.runShuffleSeed = seed
		return nil
	}
}

// RunConcurrent - Run runbooks concurrently.
func RunConcurrent(enable bool, max int) Option { //nostyle:repetition
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.runConcurrent = enable
		bk.runConcurrentMax = max
		return nil
	}
}

// RunRandom - Run the specified number of runbooks at random. Sometimes the same runbook is run multiple times.
func RunRandom(n int) Option { //nostyle:repetition
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		if n <= 0 {
			return fmt.Errorf("ramdom must be greater than 0: %d", n)
		}
		bk.runRandom = n
		return nil
	}
}

// Stdout - Set STDOUT.
func Stdout(w io.Writer) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.stdout = w
		return nil
	}
}

// Stderr - Set STDERR.
func Stderr(w io.Writer) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.stderr = w
		return nil
	}
}

// LoadOnly - Load only.
func LoadOnly() Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.loadOnly = true
		return nil
	}
}

// Scopes - Set scopes for runn.
func Scopes(scopes ...string) Option {
	return func(bk *book) error {
		return setScopes(scopes...)
	}
}

// bookWithStore - Load runbook with store.
func bookWithStore(path string, store map[string]any) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		loaded, err := loadBook(path, store)
		if err != nil {
			return err
		}
		return bk.merge(loaded)
	}
}

// setupBuiltinFunctions - Set up built-in functions to runner.
func setupBuiltinFunctions(opts ...Option) []Option {
	// Built-in functions are added at the beginning of an option and are overridden by subsequent options
	return append([]Option{
		// NOTE: Please add here the built-in functions you want to enable.
		Func("url", func(v string) *url.URL { return builtin.Url(v) }),
		Func("urlencode", url.QueryEscape),
		Func("base64encode", func(v any) string { panic("base64encode() is deprecated. Use toBase64() instead.") }),
		Func("base64decode", func(v any) string { panic("base64decode() is deprecated. Use fromBase64() instead.") }),
		Func("bool", func(v any) bool { return cast.ToBool(v) }),
		Func("time", builtin.Time),
		Func("compare", builtin.Compare),
		Func("diff", builtin.Diff),
		Func("intersect", builtin.Intersect),
		Func("input", func(msg, defaultMsg any) string {
			return prompter.Prompt(cast.ToString(msg), cast.ToString(defaultMsg))
		}),
		Func("secret", func(msg any) string {
			return prompter.Password(cast.ToString(msg))
		}),
		Func("select", func(msg any, list []any, defaultSelect any) string {
			var choices []string
			for _, v := range list {
				choices = append(choices, cast.ToString(v))
			}
			return prompter.Choose(cast.ToString(msg), choices, cast.ToString(defaultSelect))
		}),
		Func("basename", filepath.Base),
		Func("faker", builtin.NewFaker()),
		Func("json", builtin.NewJSON()),
	},
		opts...,
	)
}

func included(included bool) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.included = included
		return nil
	}
}

// Books - Load multiple runbooks.
func Books(pathp string) ([]Option, error) {
	paths, err := fetchPaths(pathp)
	if err != nil {
		return nil, err
	}
	var opts []Option
	for _, p := range paths {
		opts = append(opts, Book(p))
	}
	return opts, nil
}

func runnHTTPRunner(name string, r *httpRunner) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.httpRunners[name] = r
		return nil
	}
}

func runnDBRunner(name string, r *dbRunner) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.dbRunners[name] = r
		return nil
	}
}

func runnGrpcRunner(name string, r *grpcRunner) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.grpcRunners[name] = r
		return nil
	}
}

func runnSSHRunner(name string, r *sshRunner) Option {
	return func(bk *book) error {
		if bk == nil {
			return ErrNilBook
		}
		bk.sshRunners[name] = r
		return nil
	}
}

var (
	AsTestHelper = T
	Runbook      = Book
	RunPart      = RunShard //nostyle:repetition
)
