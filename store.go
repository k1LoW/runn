package runn

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
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
)

const (
	storeStepKeyRun     = "run"
	storeStepKeyOutcome = "outcome"
)

const (
	storeRunnKeyKV = "kv"
	storeFuncValue = "[func]"
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
}

type store struct {
	steps       []map[string]any
	stepMapKeys []string
	stepMap     map[string]map[string]any
	vars        map[string]any
	funcs       map[string]any
	bindVars    map[string]any
	parentVars  map[string]any
	useMap      bool // Use map syntax in `steps:`.
	loopIndex   *int
	cookies     map[string]map[string]*http.Cookie
	kv          *kv
}

func (s *store) recordAsMapped(k string, v map[string]any) {
	if !s.useMap {
		panic("recordAsMapped can only be used if useMap = true")
	}
	s.stepMap[k] = v
	s.stepMapKeys = append(s.stepMapKeys, k)
}

func (s *store) removeLatestAsMapped() {
	if !s.useMap {
		panic("removeLatestAsMapped can only be used if useMap = true")
	}
	latestKey := s.stepMapKeys[len(s.stepMapKeys)-1]
	delete(s.stepMap, latestKey)
	s.stepMapKeys = s.stepMapKeys[:len(s.stepMapKeys)-1]
}

func (s *store) recordAsListed(v map[string]any) {
	if s.useMap {
		panic("recordAsMapped can only be used if useMap = false")
	}
	s.steps = append(s.steps, v)
}

func (s *store) length() int {
	if s.useMap {
		return len(s.stepMapKeys)
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
	if len(s.stepMapKeys) < 2 {
		return nil
	}
	pk := s.stepMapKeys[len(s.stepMapKeys)-2]
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
	if len(s.stepMapKeys) == 0 {
		return nil
	}
	lk := s.stepMapKeys[len(s.stepMapKeys)-1]
	if v, ok := s.stepMap[lk]; ok {
		return v
	}
	return nil
}

func (s *store) recordToLatest(key string, value any) error {
	if !s.useMap {
		if len(s.steps) == 0 {
			return errors.New("failed to record")
		}
		s.steps[len(s.steps)-1][key] = value
		return nil
	}
	if len(s.stepMapKeys) == 0 {
		return errors.New("failed to record")
	}
	lk := s.stepMapKeys[len(s.stepMapKeys)-1]
	if _, ok := s.stepMap[lk]; ok {
		s.stepMap[lk][key] = value
		return nil
	}
	return errors.New("failed to record")
}

func (s *store) recordToCookie(cookies []*http.Cookie) {
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

func (s *store) toNormalizedMap() map[string]any {
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

	// runn.kv
	runnm := map[string]any{}
	if s.kv != nil {
		s.kv.mu.Lock()
		kv := map[string]any{}
		for k, v := range s.kv.m {
			kv[k] = v
		}
		runnm[storeRunnKeyKV] = kv
		s.kv.mu.Unlock()
	}
	store[storeRootKeyRunn] = runnm

	return store
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
	if s.parentVars != nil {
		store[storeRootKeyParent] = s.parentVars
	} else {
		store[storeRootKeyParent] = nil
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

	// runn.kv
	runnm := map[string]any{}
	if s.kv != nil {
		s.kv.mu.Lock()
		kv := map[string]any{}
		for k, v := range s.kv.m {
			kv[k] = v
		}
		runnm[storeRunnKeyKV] = kv
		s.kv.mu.Unlock()
	}
	store[storeRootKeyRunn] = runnm

	return store
}

func (s *store) toMapWithOutFuncs() map[string]any {
	store := map[string]any{}
	store[storeRootKeyEnv] = envMap()
	store[storeRootKeyVars] = s.vars
	if s.useMap {
		store[storeRootKeySteps] = s.stepMap
	} else {
		store[storeRootKeySteps] = s.steps
	}
	if s.parentVars != nil {
		store[storeRootKeyParent] = s.parentVars
	} else {
		store[storeRootKeyParent] = nil
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

	// runn.kv
	runnm := map[string]any{}
	if s.kv != nil {
		s.kv.mu.Lock()
		kv := map[string]any{}
		for k, v := range s.kv.m {
			kv[k] = v
		}
		runnm[storeRunnKeyKV] = kv
		s.kv.mu.Unlock()
	}
	store[storeRootKeyRunn] = runnm

	return store
}

func (s *store) clearSteps() {
	s.steps = []map[string]any{}
	s.stepMapKeys = []string{}
	s.stepMap = map[string]map[string]any{}
	// keep vars, bindVars, cookies, kv, parentVars

	s.loopIndex = nil
}

func envMap() map[string]string {
	m := map[string]string{}
	for _, e := range os.Environ() {
		splitted := strings.SplitN(e, "=", 2)
		m[splitted[0]] = splitted[1]
	}
	return m
}
