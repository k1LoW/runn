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
	"github.com/k1LoW/sshc/v4"
)

const noDesc = "[No Description]"

// book - Aggregated settings. runbook settings and run settings are aggregated.
type book struct {
	desc             string
	runners          map[string]any
	vars             map[string]any
	rawSteps         []map[string]any
	debug            bool
	ifCond           string
	skipTest         bool
	funcs            map[string]any
	stepKeys         []string
	path             string // runbook file path
	httpRunners      map[string]*httpRunner
	dbRunners        map[string]*dbRunner
	grpcRunners      map[string]*grpcRunner
	cdpRunners       map[string]*cdpRunner
	sshRunners       map[string]*sshRunner
	profile          bool
	intervalStr      string
	interval         time.Duration
	loop             *Loop
	concurrency      string
	useMap           bool
	t                *testing.T
	included         bool
	force            bool
	failFast         bool
	skipIncluded     bool
	grpcNoTLS        bool
	grpcProtos       []string
	grpcImportPaths  []string
	runMatch         *regexp.Regexp
	runSample        int
	runShardIndex    int
	runShardN        int
	runShuffle       bool
	runShuffleSeed   int64
	runConcurrent    bool
	runConcurrentMax int
	runRandom        int
	runnerErrs       map[string]error
	beforeFuncs      []func(*RunResult) error
	afterFuncs       []func(*RunResult) error
	capturers        capturers
	stdout           io.Writer
	stderr           io.Writer
	// skip some errors for `runn list`
	loadOnly bool
}

func LoadBook(path string) (*book, error) {
	return loadBook(path, nil)
}

func loadBook(path string, store map[string]any) (*book, error) {
	fp, err := fetchPath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}
	f, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}
	bk, err := parseBook(f)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
	}
	bk.path = fp
	if err := bk.parseRunners(store); err != nil {
		return nil, err
	}
	if err := bk.parseVars(store); err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("failed to load runbook %s: %w", path, err)
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
	notSSHRunners := []string{}
	if store != nil {
		r, err := EvalExpand(bk.runners, store)
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
		v, err := EvalExpand(bk.vars, store)
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

		// SSH Runner
		if !detect {
			detect, err = bk.parseSSHRunnerWithDetailed(k, tmp)
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
	if c.OpenApi3DocLocation != "" && !strings.HasPrefix(c.OpenApi3DocLocation, "https://") && !strings.HasPrefix(c.OpenApi3DocLocation, "http://") && !strings.HasPrefix(c.OpenApi3DocLocation, "/") {
		c.OpenApi3DocLocation = fp(c.OpenApi3DocLocation, root)
	}
	if c.CACert != "" {
		b, err := readFile(fp(c.CACert, root))
		if err != nil {
			return false, err
		}
		r.cacert = b
	}
	if c.Cert != "" {
		b, err := readFile(fp(c.Cert, root))
		if err != nil {
			return false, err
		}
		r.cert = b
	}
	if c.Key != "" {
		b, err := readFile(fp(c.Key, root))
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
	if c.cacert != nil {
		r.cacert = c.cacert
	} else if c.CACert != "" {
		b, err := readFile(fp(c.CACert, root))
		if err != nil {
			return false, err
		}
		r.cacert = b
	}
	if c.cert != nil {
		r.cert = c.cert
	} else if c.Cert != "" {
		b, err := readFile(fp(c.Cert, root))
		if err != nil {
			return false, err
		}
		r.cert = b
	}
	if c.key != nil {
		r.key = c.key
	} else if c.Key != "" {
		b, err := readFile(fp(c.Key, root))
		if err != nil {
			return false, err
		}
		r.key = b
	}
	r.skipVerify = c.SkipVerify
	r.importPaths = c.ImportPaths
	r.protos = c.Protos
	bk.grpcRunners[name] = r
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
	opts := []sshc.Option{}
	if c.SSHConfig != "" {
		p := c.SSHConfig
		if !strings.HasPrefix(c.SSHConfig, "/") {
			p = filepath.Join(root, c.SSHConfig)
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
		p := c.IdentityFile
		if !strings.HasPrefix(c.IdentityFile, "/") {
			p = filepath.Join(root, c.IdentityFile)
		}
		b, err := readFile(p)
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
			return false, fmt.Errorf("invalid SSH runner: '%s': invalid localForward option: %s", name, c.LocalForward)
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
		return false, err
	}
	r := &sshRunner{
		name:         name,
		client:       client,
		keepSession:  c.KeepSession,
		localForward: lf,
	}

	if r.keepSession {
		if err := r.startSession(); err != nil {
			return false, err
		}
	}

	bk.sshRunners[name] = r
	return true, nil
}

func (bk *book) applyOptions(opts ...Option) error {
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
	for k, v := range loaded.vars {
		bk.vars[k] = v
	}
	bk.runnerErrs = loaded.runnerErrs
	bk.rawSteps = loaded.rawSteps
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
	bk.loop = loaded.loop
	bk.grpcNoTLS = loaded.grpcNoTLS
	bk.grpcProtos = loaded.grpcProtos
	bk.grpcImportPaths = loaded.grpcImportPaths
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

func fp(p, root string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return filepath.Join(root, p)
}

func newBook() *book {
	return &book{
		runners:     map[string]any{},
		vars:        map[string]any{},
		rawSteps:    []map[string]any{},
		funcs:       map[string]any{},
		httpRunners: map[string]*httpRunner{},
		dbRunners:   map[string]*dbRunner{},
		grpcRunners: map[string]*grpcRunner{},
		cdpRunners:  map[string]*cdpRunner{},
		sshRunners:  map[string]*sshRunner{},
		interval:    0 * time.Second,
		runnerErrs:  map[string]error{},
		stdout:      os.Stdout,
		stderr:      os.Stderr,
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
	if k == includeRunnerKey || k == testRunnerKey || k == dumpRunnerKey || k == execRunnerKey || k == bindRunnerKey {
		return fmt.Errorf("runner name '%s' is reserved for built-in runner", k)
	}
	if k == ifSectionKey || k == descSectionKey || k == loopSectionKey {
		return fmt.Errorf("runner name '%s' is reserved for built-in section", k)
	}
	return nil
}

func validateStepKeys(s map[string]any) error {
	if len(s) == 0 {
		return errors.New("step must specify at least one runner")
	}
	custom := 0
	for k := range s {
		if k == testRunnerKey || k == dumpRunnerKey || k == bindRunnerKey || k == ifSectionKey || k == descSectionKey || k == loopSectionKey {
			continue
		}
		custom += 1
	}
	if custom > 1 {
		return errors.New("runners that cannot be running at the same time are specified")
	}
	return nil
}

func repairKey(in string) string {
	repairRep := strings.NewReplacer("-----BEGIN OPENSSH PRIVATE KEY-----", "-----BEGIN_OPENSSH_PRIVATE_KEY-----", "-----END OPENSSH PRIVATE KEY-----", "-----END_OPENSSH_PRIVATE_KEY-----", " ", "\n", "-----BEGIN_OPENSSH_PRIVATE_KEY-----", "-----BEGIN OPENSSH PRIVATE KEY-----", "-----END_OPENSSH_PRIVATE_KEY-----", "-----END OPENSSH PRIVATE KEY-----")
	return repairRep.Replace(repairRep.Replace(in))
}
