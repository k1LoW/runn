package runn

import (
	"errors"
	"os"
	"strings"
)

const (
	storeVarsKey     = "vars"
	storeStepsKey    = "steps"
	storeParentKey   = "parent"
	storeIncludedKey = "included"
	storeCurrentKey  = "current"
	storePreviousKey = "previous"
	storeEnvKey      = "env"
	storeFuncValue   = "[func]"
	storeStepRunKey  = "run"
	storeOutcomeKey  = "outcome"
)

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
}

func (s *store) recordAsMapped(k string, v map[string]any) {
	if !s.useMap {
		panic("recordAsMapped can only be used if useMap = true")
	}
	s.stepMap[k] = v
	s.stepMapKeys = append(s.stepMapKeys, k)
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

func (s *store) toNormalizedMap() map[string]any {
	store := map[string]any{}
	store[storeEnvKey] = envMap()
	for k := range s.funcs {
		store[k] = storeFuncValue
	}
	store[storeVarsKey] = s.vars
	if s.useMap {
		store[storeStepsKey] = s.stepMap
	} else {
		store[storeStepsKey] = s.steps
	}
	for k, v := range s.bindVars {
		store[k] = v
	}
	if s.loopIndex != nil {
		store[loopCountVarKey] = *s.loopIndex
	}
	return store
}

func (s *store) toMap() map[string]any {
	store := map[string]any{}
	store[storeEnvKey] = envMap()
	for k, v := range s.funcs {
		store[k] = v
	}
	store[storeVarsKey] = s.vars
	if s.useMap {
		store[storeStepsKey] = s.stepMap
	} else {
		store[storeStepsKey] = s.steps
	}
	if s.parentVars != nil {
		store[storeParentKey] = s.parentVars
	}
	for k, v := range s.bindVars {
		store[k] = v
	}
	if s.loopIndex != nil {
		store[loopCountVarKey] = *s.loopIndex
	}
	return store
}

func (s *store) clearSteps() {
	s.steps = []map[string]any{}
	s.stepMapKeys = []string{}
	s.stepMap = map[string]map[string]any{}
	// keep vars, bindVars
	s.parentVars = map[string]any{}
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
