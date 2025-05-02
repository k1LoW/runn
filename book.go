package runn

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/duration"
	"github.com/k1LoW/runn/internal/expr"
	"github.com/k1LoW/runn/internal/fs"
	"github.com/k1LoW/sshc/v4"
)

const noDesc = "[No Description]"

// book - Aggregated settings. runbook settings and run settings are aggregated.
type book struct {
	desc                 string
	labels               []string
	needs                map[string]string
	runners              map[string]any
	vars                 map[string]any
	secrets              []string
	rawSteps             []map[string]any
	hostRules            hostRules
	hostRulesFromOpts    hostRules
	debug                bool
	ifCond               string
	skipTest             bool
	funcs                map[string]any
	stepKeys             []string
	path                 string // runbook file path
	httpRunners          map[string]*httpRunner
	dbRunners            map[string]*dbRunner
	grpcRunners          map[string]*grpcRunner
	cdpRunners           map[string]*cdpRunner
	sshRunners           map[string]*sshRunner
	includeRunners       map[string]*includeRunner
	profile              bool
	intervalStr          string
	interval             time.Duration
	loop                 *Loop
	concurrency          []string
	useMap               bool
	t                    *testing.T
	included             bool
	force                bool
	trace                bool
	attach               bool
	waitTimeout          time.Duration // waitTimout is the time to wait for sub-processes to complete after the Run or RunN context is canceled
	failFast             bool
	skipIncluded         bool
	openAPI3DocLocations []string
	grpcNoTLS            bool
	grpcProtos           []string
	grpcImportPaths      []string
	grpcBufDirs          []string
	grpcBufLocks         []string
	grpcBufConfigs       []string
	grpcBufModules       []string
	runIDs               []string
	runMatch             *regexp.Regexp
	runLabels            []string
	runSample            int
	runShardIndex        int
	runShardN            int
	runShuffle           bool
	runShuffleSeed       int64
	runConcurrent        bool
	runConcurrentMax     int
	runRandom            int
	runnerErrs           map[string]error
	beforeFuncs          []func(*RunResult) error
	afterFuncs           []func(*RunResult) error
	capturers            capturers
	stdout               io.Writer
	stderr               io.Writer
	// Skip some errors for `runn list`
	loadOnly bool
}

func LoadBook(path string) (*book, error) {
	return loadBook(path, nil)
}

func loadBook(path string, store map[string]any) (_ *book, err error) {
	fp, err := fs.FetchPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}
	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}
	defer func() {
		if errr := f.Close(); errr != nil {
			err = errors.Join(err, fmt.Errorf("failed to load runbook %s: %w", path, errr))
		}
	}()
	bk, err := parseBook(f)
	if err != nil {
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}
	bk.path = fp
	if err := bk.parseRunners(store); err != nil {
		return nil, err
	}
	if err := bk.parseVars(store); err != nil {
		return nil, err
	}

	return bk, nil
}

func (bk *book) Desc() string {
	return bk.desc
}

func (bk *book) If() string {
	return bk.ifCond
}

func (bk *book) parseRunners(store map[string]any) error {
	// parse SSH Runners first for port forwarding
	var notSSHRunners []string
	if store != nil {
		r, err := expr.EvalExpand(bk.runners, store)
		if err != nil {
			return err
		}
		var ok bool
		bk.runners, ok = r.(map[string]any)
		if !ok {
			return fmt.Errorf("failed to cast: %v", r)
		}
	}
	for k, v := range bk.runners {
		if detectSSHRunner(v) {
			if err := bk.parseRunner(k, v); err != nil {
				bk.runnerErrs[k] = err
			}
			continue
		}
		notSSHRunners = append(notSSHRunners, k)
	}
	for _, k := range notSSHRunners {
		v := bk.runners[k]
		if err := bk.parseRunner(k, v); err != nil {
			bk.runnerErrs[k] = err
		}
	}
	return nil
}

func (bk *book) parseVars(store map[string]any) error {
	if store != nil {
		v, err := expr.EvalExpand(bk.vars, store)
		if err != nil {
			return err
		}
		var ok bool
		bk.vars, ok = v.(map[string]any)
		if !ok {
			return fmt.Errorf("failed to cast: %v", v)
		}
	}
	for k, v := range bk.vars {
		root, err := bk.generateOperatorRoot()
		if err != nil {
			return err
		}
		ev, err := evaluateSchema(v, root, store)
		if err != nil {
			return err
		}
		bk.vars[k] = ev
	}
	return nil
}

func (bk *book) parseRunner(k string, v any) error {
	delete(bk.runnerErrs, k)

	switch vv := v.(type) {
	case string:
		switch {
		case strings.HasPrefix(vv, "https://") || strings.HasPrefix(vv, "http://"):
			hc, err := newHTTPRunner(k, vv)
			if err != nil {
				return err
			}
			bk.httpRunners[k] = hc
		case strings.HasPrefix(vv, "grpc://"):
			addr := strings.TrimPrefix(vv, "grpc://")
			gc, err := newGrpcRunner(k, addr)
			if err != nil {
				return err
			}
			bk.grpcRunners[k] = gc
		case strings.HasPrefix(vv, "cdp://") || strings.HasPrefix(vv, "chrome://"):
			remote := strings.TrimPrefix(strings.TrimPrefix(vv, "cdp://"), "chrome://")
			cc, err := newCDPRunner(k, remote)
			if err != nil {
				return err
			}
			bk.cdpRunners[k] = cc
		case strings.HasPrefix(vv, "ssh://"):
			addr := strings.TrimPrefix(vv, "ssh://")
			sc, err := newSSHRunner(k, addr)
			if err != nil {
				return err
			}
			bk.sshRunners[k] = sc
		default:
			dc, err := newDBRunner(k, vv)
			if err != nil {
				return err
			}
			bk.dbRunners[k] = dc
		}
	case map[string]any:
		tmp, err := yaml.Marshal(vv)
		if err != nil {
			return err
		}
		detect := false

		// HTTP Runner
		detect, err = bk.parseHTTPRunnerWithDetailed(k, tmp)
		if err != nil {
			return err
		}

		// gRPC Runner
		if !detect {
			detect, err = bk.parseGRPCRunnerWithDetailed(k, tmp)
			if err != nil {
				return err
			}
		}

		// DB Runner
		if !detect {
			detect, err = bk.parseDBRunnerWithDetailed(k, tmp)
			if err != nil {
				return err
			}
		}

		// SSH Runner
		if !detect {
			detect, err = bk.parseSSHRunnerWithDetailed(k, tmp)
			if err != nil {
				return err
			}
		}

		// Include Runner
		if !detect {
			detect, err = bk.parseIncludeRunnerWithDetailed(k, tmp)
			if err != nil {
				return err
			}
		}

		if !detect {
			return fmt.Errorf("cannot detect runner: %s", string(tmp))
		}
	}

	return nil
}

func (bk *book) parseHTTPRunnerWithDetailed(name string, b []byte) (bool, error) {
	c := &httpRunnerConfig{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return false, nil
	}
	if c.Endpoint == "" {
		return false, nil
	}
	root, err := bk.generateOperatorRoot()
	if err != nil {
		return false, err
	}
	r, err := newHTTPRunner(name, c.Endpoint)
	if err != nil {
		return false, err
	}
	bk.httpRunners[name] = r

	if c.NotFollowRedirect {
		r.client.CheckRedirect = notFollowRedirectFn
	}
	r.multipartBoundary = c.MultipartBoundary
	if c.OpenAPI3DocLocation != "" && !strings.HasPrefix(c.OpenAPI3DocLocation, "https://") && !strings.HasPrefix(c.OpenAPI3DocLocation, "http://") && !strings.HasPrefix(c.OpenAPI3DocLocation, "/") {
		c.OpenAPI3DocLocation, err = fs.Path(c.OpenAPI3DocLocation, root)
		if err != nil {
			return false, err
		}
	}
	if c.CACert != "" {
		p, err := fs.Path(c.CACert, root)
		if err != nil {
			return false, err
		}
		b, err := fs.ReadFile(p)
		if err != nil {
			return false, err
		}
		r.cacert = b
	}
	if c.Cert != "" {
		p, err := fs.Path(c.Cert, root)
		if err != nil {
			return false, err
		}
		b, err := fs.ReadFile(p)
		if err != nil {
			return false, err
		}
		r.cert = b
	}
	if c.Key != "" {
		p, err := fs.Path(c.Key, root)
		if err != nil {
			return false, err
		}
		b, err := fs.ReadFile(p)
		if err != nil {
			return false, err
		}
		r.key = b
	}
	r.skipVerify = c.SkipVerify
	if c.Timeout != "" {
		r.client.Timeout, err = duration.Parse(c.Timeout)
		if err != nil {
			return false, fmt.Errorf("timeout in HttpRunnerConfig is invalid: %w", err)
		}
	}
	r.useCookie = c.UseCookie
	r.trace = c.Trace.Enable
	r.traceHeaderName = c.Trace.HeaderName
	hv, err := newHttpValidator(c)
	if err != nil {
		return false, err
	}
	r.validator = hv
	return true, nil
}

func (bk *book) parseGRPCRunnerWithDetailed(name string, b []byte) (bool, error) {
	c := &grpcRunnerConfig{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return false, nil
	}
	if c.Addr == "" {
		return false, nil
	}
	root, err := bk.generateOperatorRoot()
	if err != nil {
		return false, err
	}
	r, err := newGrpcRunner(name, c.Addr)
	if err != nil {
		return false, err
	}
	r.tls = c.TLS
	if len(c.cacert) != 0 {
		r.cacert = c.cacert
	} else if c.CACert != "" {
		p, err := fs.Path(c.CACert, root)
		if err != nil {
			return false, err
		}
		b, err := fs.ReadFile(p)
		if err != nil {
			return false, err
		}
		r.cacert = b
	}
	if len(c.cert) != 0 {
		r.cert = c.cert
	} else if c.Cert != "" {
		p, err := fs.Path(c.Cert, root)
		if err != nil {
			return false, err
		}
		b, err := fs.ReadFile(p)
		if err != nil {
			return false, err
		}
		r.cert = b
	}
	if len(c.key) != 0 {
		r.key = c.key
	} else if c.Key != "" {
		p, err := fs.Path(c.Key, root)
		if err != nil {
			return false, err
		}
		b, err := fs.ReadFile(p)
		if err != nil {
			return false, err
		}
		r.key = b
	}
	r.skipVerify = c.SkipVerify
	for _, p := range c.ImportPaths {
		pp, err := fs.Path(p, root)
		if err != nil {
			return false, err
		}
		r.importPaths = append(r.importPaths, pp)
	}
	for _, p := range c.Protos {
		pp, err := fs.Path(p, root)
		if err != nil {
			return false, err
		}
		r.protos = append(r.protos, pp)
	}
	for _, p := range c.BufDirs {
		pp, err := fs.Path(p, root)
		if err != nil {
			return false, err
		}
		r.bufDirs = append(r.bufDirs, pp)
	}
	for _, p := range c.BufLocks {
		pp, err := fs.Path(p, root)
		if err != nil {
			return false, err
		}
		r.bufLocks = append(r.bufLocks, pp)
	}
	for _, p := range c.BufConfigs {
		pp, err := fs.Path(p, root)
		if err != nil {
			return false, err
		}
		r.bufConfigs = append(r.bufConfigs, pp)
	}
	r.bufModules = c.BufModules
	r.trace = c.Trace.Enable
	r.traceHeaderName = c.Trace.HeaderName

	bk.grpcRunners[name] = r
	return true, nil
}

func (bk *book) parseDBRunnerWithDetailed(name string, b []byte) (bool, error) {
	c := &dbRunnerConfig{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return false, nil
	}
	if c.DSN == "" {
		return false, nil
	}
	r, err := newDBRunner(name, c.DSN)
	if err != nil {
		return false, err
	}
	r.trace = c.Trace
	bk.dbRunners[name] = r
	return true, nil
}

func (bk *book) parseSSHRunnerWithDetailed(name string, b []byte) (bool, error) {
	c := &sshRunnerConfig{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return false, nil
	}
	if c.Host == "" && c.Hostname == "" {
		return false, nil
	}
	if err := c.validate(); err != nil {
		return false, err
	}
	host := c.Host
	if host == "" {
		host = c.Hostname
	}
	root, err := bk.generateOperatorRoot()
	if err != nil {
		return false, err
	}
	var opts []sshc.Option
	if c.SSHConfig != "" {
		p, err := fs.Path(c.SSHConfig, root)
		if err != nil {
			return false, err
		}
		if _, err := os.Stat(p); err != nil {
			return false, err
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
		p, err := fs.Path(c.IdentityFile, root)
		if err != nil {
			return false, err
		}
		b, err := fs.ReadFile(p)
		if err != nil {
			return false, err
		}
		opts = append(opts, sshc.IdentityKey(b))
	} else if c.IdentityKey != "" {
		opts = append(opts, sshc.IdentityKey([]byte(repairKey(c.IdentityKey))))
	}
	var lf *sshLocalForward
	if c.LocalForward != "" {
		c.KeepSession = true
		if strings.Count(c.LocalForward, ":") != 2 {
			return false, fmt.Errorf("invalid SSH runner: %q: invalid localForward option: %s", name, c.LocalForward)
		}
		splitted := strings.SplitN(c.LocalForward, ":", 2)
		lf = &sshLocalForward{
			local:  fmt.Sprintf("127.0.0.1:%s", splitted[0]),
			remote: splitted[1],
		}
	}
	opts = append(opts, sshc.AuthMethod(sshKeyboardInteractive(c.KeyboardInteractive)))

	r := &sshRunner{
		name:         name,
		addr:         host,
		keepSession:  c.KeepSession,
		localForward: lf,
		opts:         opts,
	}

	if r.keepSession {
		client, err := sshc.NewClient(host, opts...)
		if err != nil {
			return false, err
		}
		r.client = client
		if err := r.startSession(); err != nil {
			return false, err
		}
	}

	bk.sshRunners[name] = r
	return true, nil
}

func (bk *book) parseIncludeRunnerWithDetailed(name string, b []byte) (bool, error) {
	c := &includeRunnerConfig{}
	if err := yaml.Unmarshal(b, c); err != nil {
		return false, nil
	}
	if c.Path == "" {
		return false, nil
	}
	r := &includeRunner{
		name:   name,
		path:   c.Path,
		params: c.Params,
	}

	bk.includeRunners[name] = r
	return true, nil
}

func (bk *book) applyOptions(opts ...Option) error {
	// First, execute Scopes()
	for _, opt := range opts {
		_ = opt(nil)
	}
	opts = setupBuiltinFunctions(opts...)
	for _, opt := range opts {
		if err := opt(bk); err != nil {
			return err
		}
	}
	return nil
}

// generateOperatorRoot generates the root path of the operator.
func (bk *book) generateOperatorRoot() (string, error) {
	if bk.path != "" {
		return filepath.Dir(bk.path), nil
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return wd, nil
	}
}

func (bk *book) merge(loaded *book) error {
	bk.path = loaded.path
	bk.desc = loaded.desc
	bk.labels = loaded.labels
	bk.needs = loaded.needs
	bk.ifCond = loaded.ifCond
	bk.useMap = loaded.useMap
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
	for k, r := range loaded.includeRunners {
		bk.includeRunners[k] = r
	}
	for k, v := range loaded.vars {
		bk.vars[k] = v
	}
	bk.secrets = append(bk.secrets, loaded.secrets...)
	bk.runnerErrs = loaded.runnerErrs
	bk.rawSteps = loaded.rawSteps
	bk.hostRules = loaded.hostRules
	bk.hostRulesFromOpts = loaded.hostRulesFromOpts
	bk.stepKeys = loaded.stepKeys
	if !bk.debug {
		bk.debug = loaded.debug
	}
	if !bk.skipTest {
		bk.skipTest = loaded.skipTest
	}
	if !bk.force {
		bk.force = loaded.force
	}
	if !bk.trace {
		bk.trace = loaded.trace
	}
	bk.loop = loaded.loop
	bk.concurrency = loaded.concurrency
	bk.openAPI3DocLocations = loaded.openAPI3DocLocations
	bk.grpcNoTLS = loaded.grpcNoTLS
	bk.grpcProtos = loaded.grpcProtos
	bk.grpcImportPaths = loaded.grpcImportPaths
	bk.grpcBufDirs = loaded.grpcBufDirs
	bk.grpcBufLocks = loaded.grpcBufLocks
	bk.grpcBufConfigs = loaded.grpcBufConfigs
	bk.grpcBufModules = loaded.grpcBufModules
	if loaded.intervalStr != "" {
		bk.interval = loaded.interval
	}
	return nil
}

func detectSSHRunner(v any) bool {
	switch vv := v.(type) {
	case string:
		if strings.HasPrefix(vv, "ssh://") {
			return true
		}
	case map[string]any:
		b, err := yaml.Marshal(vv)
		if err != nil {
			return false
		}
		c := &sshRunnerConfig{}
		if err := yaml.Unmarshal(b, c); err != nil {
			return false
		}
		if c.Host == "" && c.Hostname == "" {
			return false
		}
		return true
	}
	return false
}

func newBook() *book {
	return &book{
		runners:        map[string]any{},
		vars:           map[string]any{},
		rawSteps:       []map[string]any{},
		funcs:          map[string]any{},
		httpRunners:    map[string]*httpRunner{},
		dbRunners:      map[string]*dbRunner{},
		grpcRunners:    map[string]*grpcRunner{},
		cdpRunners:     map[string]*cdpRunner{},
		sshRunners:     map[string]*sshRunner{},
		includeRunners: map[string]*includeRunner{},
		interval:       0 * time.Second,
		runnerErrs:     map[string]error{},
		stdout:         os.Stdout,
		stderr:         os.Stderr,
	}
}

func parseBook(in io.Reader) (*book, error) {
	rb, err := ParseRunbook(in)
	if err != nil {
		return nil, err
	}
	bk, err := rb.toBook()
	if err != nil {
		return nil, err
	}

	// To match behavior with json.Marshal
	{
		b, err := json.Marshal(bk.vars)
		if err != nil {
			return nil, fmt.Errorf("invalid vars: %w", err)
		}
		if err := json.Unmarshal(b, &bk.vars); err != nil {
			return nil, fmt.Errorf("invalid vars: %w", err)
		}
	}

	if bk.desc == "" {
		bk.desc = noDesc
	}

	if bk.intervalStr != "" {
		d, err := parseDuration(bk.intervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid interval: %w", err)
		}
		bk.interval = d
	}

	for k := range bk.runners {
		if err := validateRunnerKey(k); err != nil {
			return nil, err
		}
	}

	for i, s := range bk.rawSteps {
		if err := validateStepKeys(s); err != nil {
			return nil, fmt.Errorf("invalid steps[%d]. %w: %s", i, err, s)
		}
	}

	return bk, nil
}

func validateRunnerKey(k string) error {
	if k == includeRunnerKey || k == testRunnerKey || k == dumpRunnerKey || k == execRunnerKey || k == bindRunnerKey || k == runnerRunnerKey {
		return fmt.Errorf("runner name %q is reserved for built-in runner", k)
	}
	if k == ifSectionKey || k == descSectionKey || k == loopSectionKey || k == deferSectionKey || k == forceSectionKey {
		return fmt.Errorf("runner name %q is reserved for built-in section", k)
	}
	return nil
}

func validateStepKeys(s map[string]any) error {
	if len(s) == 0 {
		return errors.New("step must specify at least one runner")
	}
	var mainRunnerKey string
	mainRunner := 0
	subRunner := 0
	for k := range s {
		if k == ifSectionKey || k == descSectionKey || k == loopSectionKey || k == deferSectionKey || k == forceSectionKey {
			continue
		}
		if k == testRunnerKey || k == dumpRunnerKey || k == bindRunnerKey {
			subRunner += 1
			continue
		}
		mainRunner += 1
		mainRunnerKey = k
	}
	if mainRunner > 1 {
		return errors.New("runners that cannot be running at the same time are specified")
	}
	if mainRunnerKey == runnerRunnerKey && subRunner > 0 {
		return errors.New("runner: runner cannot be used with other runners")
	}

	return nil
}

func repairKey(in string) string {
	repairRep := strings.NewReplacer("-----BEGIN OPENSSH PRIVATE KEY-----", "-----BEGIN_OPENSSH_PRIVATE_KEY-----", "-----END OPENSSH PRIVATE KEY-----", "-----END_OPENSSH_PRIVATE_KEY-----", " ", "\n", "-----BEGIN_OPENSSH_PRIVATE_KEY-----", "-----BEGIN OPENSSH PRIVATE KEY-----", "-----END_OPENSSH_PRIVATE_KEY-----", "-----END OPENSSH PRIVATE KEY-----")
	return repairRep.Replace(repairRep.Replace(in))
}
