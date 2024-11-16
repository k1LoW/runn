package runn

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/k1LoW/maskedio"
)

const (
	storeRootKeyVars     = "vars"
	storeRootKeySteps    = "steps"
	storeRootKeyParent   = "parent"
	storeRootKeyIncluded = "included"
	storeRootKeyCurrent  = "current"
	storeRootKeyPrevious = "previous"
	storeRootKeyEnv      = "env"
	storeRootKeyCookie   = "cookies"
	storeRootKeyNodes    = "nodes"
	storeRootKeyParams   = "params"
	storeRootKeyRunn     = "runn"
	storeRootKeyNeeds    = "needs"
)

const (
	storeStepKeyOutcome = "outcome"
)

const (
	storeRunnKeyKV        = "kv"
	storeRunnKeyRunNIndex = "i"
	storeFuncValue        = "[func]"
)

// Reserved store root keys.
var reservedStoreRootKeys = []string{
	storeRootKeyVars,
	storeRootKeySteps,
	storeRootKeyParent,
	storeRootKeyIncluded,
	storeRootKeyCurrent,
	storeRootKeyPrevious,
	storeRootKeyEnv,
	storeRootKeyCookie,
	storeRootKeyNodes,
	storeRootKeyParams,
	storeRootKeyLoopCountIndex,
	storeRootKeyRunn,
	storeRootKeyNeeds,
}

type store struct {
	steps       []map[string]any
	stepMapKeys []string // all keys of mapped runbook
	stepMap     map[string]map[string]any
	vars        map[string]any
	funcs       map[string]any
	bindVars    map[string]any
	parentVars  map[string]any
	needsVars   map[string]any
	useMap      bool // Use map syntax in `steps:`.
	loopIndex   *int
	cookies     map[string]map[string]*http.Cookie
	kv          *kv
	runNIndex   int

	// for secret masking
	secrets []string // Secret var names to be masked.
	mr      *maskedio.Rule
}

func newStore(vars, funcs map[string]any, secrets []string, useMap bool, stepMapKeys []string) *store {
	return &store{
		steps:       []map[string]any{},
		stepMap:     map[string]map[string]any{},
		stepMapKeys: stepMapKeys,
		vars:        vars,
		funcs:       funcs,
		bindVars:    map[string]any{},
		parentVars:  map[string]any{},
		needsVars:   map[string]any{},
		useMap:      useMap,
		runNIndex:   -1,
		secrets:     secrets,
		mr:          maskedio.NewRule(),
	}
}

func (s *store) record(v map[string]any) {
	if s.useMap {
		s.recordAsMapped(v)
	} else {
		s.recordAsListed(v)
	}
}

func (s *store) recordAsMapped(v map[string]any) {
	if !s.useMap {
		panic("recordAsMapped can only be used if useMap = true")
	}
	if s.loopIndex != nil && *s.loopIndex > 0 {
		// delete values of prevous loop
		s.removeLatestAsMapped()
	}
	k := s.stepMapKeys[s.length()]
	s.stepMap[k] = v
}

func (s *store) removeLatestAsMapped() {
	if !s.useMap {
		panic("removeLatestAsMapped can only be used if useMap = true")
	}
	latestKey := s.stepMapKeys[len(s.stepMapKeys)-1]
	delete(s.stepMap, latestKey)
}

func (s *store) recordAsListed(v map[string]any) {
	if s.useMap {
		panic("recordAsMapped can only be used if useMap = false")
	}
	if s.loopIndex != nil && *s.loopIndex > 0 {
		// delete values of prevous loop
		s.steps = s.steps[:s.length()-1]
	}
	s.steps = append(s.steps, v)
}

func (s *store) length() int {
	if s.useMap {
		return len(s.stepMap)
	}
	return len(s.steps)
}

func (s *store) previous() map[string]any {
	if !s.useMap {
		if len(s.steps) < 2 {
			return nil
		}
		return s.steps[len(s.steps)-2]
	}
	if len(s.stepMap) < 2 {
		return nil
	}
	pk := s.stepMapKeys[len(s.stepMap)-2]
	if v, ok := s.stepMap[pk]; ok {
		return v
	}
	return nil
}

func (s *store) latest() map[string]any {
	if !s.useMap {
		if len(s.steps) == 0 {
			return nil
		}
		return s.steps[len(s.steps)-1]
	}
	if len(s.stepMap) == 0 {
		return nil
	}
	lk := s.stepMapKeys[len(s.stepMap)-1]
	if v, ok := s.stepMap[lk]; ok {
		return v
	}
	return nil
}

func (s *store) recordToLatestStep(key string, value any) error {
	if !s.useMap {
		if len(s.steps) == 0 {
			return errors.New("failed to record: store.steps is zero")
		}
		s.steps[len(s.steps)-1][key] = value
		return nil
	}
	if len(s.stepMap) == 0 {
		return errors.New("failed to record: store.stepMapKeys is zero")
	}
	lk := s.stepMapKeys[len(s.stepMap)-1]
	if _, ok := s.stepMap[lk]; ok {
		s.stepMap[lk][key] = value
		return nil
	}
	return errors.New("failed to record")
}

func (s *store) recordCookie(cookies []*http.Cookie) {
	cookieMap := make(map[string]map[string]*http.Cookie)
	for _, cookie := range cookies {
		domain := cookie.Domain
		if domain == "" {
			domain = "localhost"
		}
		keyMap, ok := cookieMap[domain]
		if !ok || keyMap == nil {
			keyMap = make(map[string]*http.Cookie)
		}
		if !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now()) {
			// Remove expired cookie
			delete(keyMap, cookie.Name)
		} else {
			keyMap[cookie.Name] = cookie
		}
		cookieMap[domain] = keyMap
	}
	s.cookies = cookieMap
}

func (s *store) toMap() map[string]any {
	store := map[string]any{}
	store[storeRootKeyEnv] = envMap()
	for k, v := range s.funcs {
		store[k] = v
	}
	store[storeRootKeyVars] = s.vars
	if s.useMap {
		store[storeRootKeySteps] = s.stepMap
	} else {
		store[storeRootKeySteps] = s.steps
	}
	if len(s.parentVars) > 0 {
		store[storeRootKeyParent] = s.parentVars
	} else {
		store[storeRootKeyParent] = nil
	}
	if len(s.needsVars) > 0 {
		store[storeRootKeyNeeds] = s.needsVars
	} else {
		store[storeRootKeyNeeds] = nil
	}
	for k, v := range s.bindVars {
		store[k] = v
	}
	if s.loopIndex != nil {
		store[storeRootKeyLoopCountIndex] = *s.loopIndex
	}
	if s.cookies != nil {
		store[storeRootKeyCookie] = s.cookies
	}

	runnm := map[string]any{}
	// runn.kv
	if s.kv != nil {
		s.kv.mu.Lock()
		kv := map[string]any{}
		for k, v := range s.kv.m {
			kv[k] = v
		}
		runnm[storeRunnKeyKV] = kv
		s.kv.mu.Unlock()
	}
	// runn.i
	runnm[storeRunnKeyRunNIndex] = s.runNIndex
	store[storeRootKeyRunn] = runnm

	return store
}

// toMapForIncludeRunner - returns a map for include runner.
// toMap without s.parentVars and s.needsVars.
func (s *store) toMapForIncludeRunner() map[string]any {
	store := map[string]any{}
	store[storeRootKeyEnv] = envMap()
	for k := range s.funcs {
		store[k] = storeFuncValue
	}
	store[storeRootKeyVars] = s.vars
	if s.useMap {
		store[storeRootKeySteps] = s.stepMap
	} else {
		store[storeRootKeySteps] = s.steps
	}
	for k, v := range s.bindVars {
		store[k] = v
	}
	if s.loopIndex != nil {
		store[storeRootKeyLoopCountIndex] = *s.loopIndex
	}
	if s.cookies != nil {
		store[storeRootKeyCookie] = s.cookies
	}

	runnm := map[string]any{}
	// runn.kv
	if s.kv != nil {
		s.kv.mu.Lock()
		kv := map[string]any{}
		for k, v := range s.kv.m {
			kv[k] = v
		}
		runnm[storeRunnKeyKV] = kv
		s.kv.mu.Unlock()
	}
	// runn.i
	runnm[storeRunnKeyRunNIndex] = s.runNIndex

	store[storeRootKeyRunn] = runnm

	return store
}

// toMapForDbg - returns a map for dbg.
// toMap without s.funcs.
func (s *store) toMapForDbg() map[string]any {
	store := map[string]any{}
	store[storeRootKeyEnv] = envMap()
	store[storeRootKeyVars] = s.vars
	if s.useMap {
		store[storeRootKeySteps] = s.stepMap
	} else {
		store[storeRootKeySteps] = s.steps
	}
	if len(s.parentVars) > 0 {
		store[storeRootKeyParent] = s.parentVars
	} else {
		store[storeRootKeyParent] = nil
	}
	if len(s.needsVars) > 0 {
		store[storeRootKeyNeeds] = s.needsVars
	} else {
		store[storeRootKeyNeeds] = nil
	}
	for k, v := range s.bindVars {
		store[k] = v
	}
	if s.loopIndex != nil {
		store[storeRootKeyLoopCountIndex] = *s.loopIndex
	}
	if s.cookies != nil {
		store[storeRootKeyCookie] = s.cookies
	}

	runnm := map[string]any{}
	// runn.kv
	if s.kv != nil {
		s.kv.mu.Lock()
		kv := map[string]any{}
		for k, v := range s.kv.m {
			kv[k] = v
		}
		runnm[storeRunnKeyKV] = kv
		s.kv.mu.Unlock()
	}
	// runn.i
	runnm[storeRunnKeyRunNIndex] = s.runNIndex
	store[storeRootKeyRunn] = runnm

	return store
}

func (s *store) clearSteps() {
	s.steps = []map[string]any{}
	s.stepMap = map[string]map[string]any{}
	// keep stepMapKeys, vars, bindVars, cookies, kv, parentVars, runNIndex

	s.loopIndex = nil
}

func (s *store) maskRule() *maskedio.Rule {
	return s.mr
}

func envMap() map[string]string {
	m := map[string]string{}
	for _, e := range os.Environ() {
		splitted := strings.SplitN(e, "=", 2)
		m[splitted[0]] = splitted[1]
	}
	return m
}
