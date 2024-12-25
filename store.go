package runn

import (
	"errors"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/k1LoW/maskedio"
	"github.com/k1LoW/runn/internal/kv"
	"github.com/mattn/go-isatty"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

const (
	storeRootKeyVars           = "vars"
	storeRootKeySteps          = "steps"
	storeRootKeyParent         = "parent"
	storeRootKeyIncluded       = "included"
	storeRootKeyCurrent        = "current"
	storeRootKeyPrevious       = "previous"
	storeRootKeyEnv            = "env"
	storeRootKeyCookie         = "cookies"
	storeRootKeyNodes          = "nodes"
	storeRootKeyParams         = "params"
	storeRootKeyRunn           = "runn"
	storeRootKeyNeeds          = "needs"
	storeRootKeyLoopCountIndex = "i"
)

const (
	storeStepKeyOutcome = "outcome"
	storeFuncValue      = "[func]"
)

// `runn` is global store for runn.
const (
	storeRunnKeyKV        = "kv"
	storeRunnKeyRunNIndex = "i"
	storeRunnKeyStdin     = "stdin"
)

var stdin any

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
	stepList    map[int]map[string]any
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
	kv          *kv.KV
	runNIndex   int

	// for secret masking
	secrets []string // Secret var names to be masked.
	mr      *maskedio.Rule
}

func newStore(vars, funcs map[string]any, secrets []string, useMap bool, stepMapKeys []string) *store {
	s := &store{
		stepList:    map[int]map[string]any{},
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
	s.setMaskKeywords(s.toMap())

	return s
}

func (s *store) record(idx int, v map[string]any) {
	if s.useMap {
		s.recordAsMapped(idx, v)
	} else {
		s.recordAsListed(idx, v)
	}
}

func (s *store) recordAsMapped(idx int, v map[string]any) {
	if !s.useMap {
		panic("recordAsMapped can only be used if useMap = true")
	}
	k := s.stepMapKeys[idx]
	s.stepMap[k] = v
}

func (s *store) recordAsListed(idx int, v map[string]any) {
	if s.useMap {
		panic("recordAsMapped can only be used if useMap = false")
	}
	s.stepList[idx] = v
}

func (s *store) length() int {
	if s.useMap {
		return len(s.stepMap)
	}
	return len(s.stepList)
}

func (s *store) previous() map[string]any {
	if s.length() < 2 {
		return nil
	}
	if !s.useMap {
		keys := lo.Keys(s.stepList)
		slices.Sort(keys)
		if v, ok := s.stepList[keys[len(keys)-2]]; ok {
			return v
		}
		return nil
	}

	var idxs []int
	for k := range s.stepMap {
		for idx, key := range s.stepMapKeys {
			if key == k {
				idxs = append(idxs, idx)
			}
		}
	}
	slices.Sort(idxs)
	key := s.stepMapKeys[idxs[len(idxs)-2]]
	if v, ok := s.stepMap[key]; ok {
		return v
	}
	return nil
}

func (s *store) latest() map[string]any {
	if s.length() == 0 {
		return nil
	}
	if !s.useMap {
		keys := lo.Keys(s.stepList)
		slices.Sort(keys)
		if v, ok := s.stepList[keys[len(keys)-1]]; ok {
			return v
		}
		return nil
	}
	var idxs []int
	for k := range s.stepMap {
		for idx, key := range s.stepMapKeys {
			if key == k {
				idxs = append(idxs, idx)
			}
		}
	}
	slices.Sort(idxs)
	key := s.stepMapKeys[idxs[len(idxs)-1]]
	if v, ok := s.stepMap[key]; ok {
		return v
	}
	return nil
}

func (s *store) recordTo(idx int, key string, value any) error {
	if s.length() == 0 {
		return errors.New("failed to record: store.steps is zero")
	}
	if !s.useMap {
		if _, ok := s.stepList[idx]; ok {
			s.stepList[idx][key] = value
			return nil
		}
		return errors.New("failed to record")
	}
	lk := s.stepMapKeys[idx]
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
		store[storeRootKeySteps] = convertStepListToList(s.stepList)
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
		kv := map[string]any{}
		for _, k := range s.kv.Keys() {
			kv[k] = s.kv.Get(k)
		}
		runnm[storeRunnKeyKV] = kv
	}
	// runn.i
	runnm[storeRunnKeyRunNIndex] = s.runNIndex
	// runn.stdin
	if stdin != nil {
		runnm[storeRunnKeyStdin] = stdin
	}
	store[storeRootKeyRunn] = runnm

	s.setMaskKeywords(store)

	return store
}

// toMapForIncludeRunner - returns a map for include runner.
// toMap without s.parentVars s.needsVars and runn.* .
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
		store[storeRootKeySteps] = convertStepListToList(s.stepList)
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
	s.setMaskKeywords(store)

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
		store[storeRootKeySteps] = convertStepListToList(s.stepList)
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
		kv := map[string]any{}
		for _, k := range s.kv.Keys() {
			kv[k] = s.kv.Get(k)
		}
		runnm[storeRunnKeyKV] = kv
	}
	// runn.i
	runnm[storeRunnKeyRunNIndex] = s.runNIndex
	// runn.stdin
	if stdin != nil {
		runnm[storeRunnKeyStdin] = stdin
	}
	store[storeRootKeyRunn] = runnm

	// runn.stdin
	if stdin != nil {
		store[storeRunnKeyStdin] = stdin
	}

	s.setMaskKeywords(store)

	return store
}

func (s *store) setMaskKeywords(store map[string]any) {
	store[storeRootKeyCurrent] = s.latest()
	defer func() {
		delete(store, storeRootKeyCurrent)
	}()

	for _, key := range s.secrets {
		v, err := Eval(key, store)
		if err != nil {
			continue
		}

		switch vv := v.(type) {
		case map[string]any:
		case []any:
		default:
			s.maskRule().SetKeyword(cast.ToString(vv))
		}
	}
}

func (s *store) clearSteps() {
	s.stepList = map[int]map[string]any{}
	s.stepMap = map[string]map[string]any{}
	// keep stepMapKeys, vars, bindVars, cookies, kv, parentVars, runNIndex

	s.loopIndex = nil
}

func (s *store) maskRule() *maskedio.Rule {
	return s.mr
}

// SetStdin reads from stdin and sets the value to store.
func SetStdin(f *os.File) error {
	if isatty.IsTerminal(f.Fd()) {
		return nil
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &stdin); err != nil {
		stdin = string(b)
	}
	return nil
}

func envMap() map[string]string {
	m := map[string]string{}
	for _, e := range os.Environ() {
		splitted := strings.SplitN(e, "=", 2)
		m[splitted[0]] = splitted[1]
	}
	return m
}

func convertStepListToList(l map[int]map[string]any) []map[string]any {
	var list []map[string]any
	idxs := lo.Keys(l)
	if len(idxs) == 0 {
		return list
	}
	slices.Sort(idxs)
	latestIdx := idxs[len(idxs)-1]
	for idx := range latestIdx + 1 {
		if _, ok := l[idx]; ok {
			list = append(list, l[idx])
		} else {
			list = append(list, nil)
		}
	}
	return list
}
