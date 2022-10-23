package runn

const (
	storeVarsKey     = "vars"
	storeStepsKey    = "steps"
	storeParentKey   = "parent"
	storeIncludedKey = "included"
	storeCurrentKey  = "current"
	storePreviousKey = "previous"
	storeFuncValue   = "[func]"
)

type store struct {
	steps       []map[string]interface{}
	stepMapKeys []string
	stepMap     map[string]map[string]interface{}
	vars        map[string]interface{}
	funcs       map[string]interface{}
	bindVars    map[string]interface{}
	parentVars  map[string]interface{}
	useMap      bool // Use map syntax in `steps:`.
	loopIndex   *int
}

func (s *store) recordAsMapped(k string, v map[string]interface{}) {
	if !s.useMap {
		panic("recordAsMapped can only be used if useMap = true")
	}
	s.stepMap[k] = v
	s.stepMapKeys = append(s.stepMapKeys, k)
}

func (s *store) recordAsListed(v map[string]interface{}) {
	if s.useMap {
		panic("recordAsMapped can only be used if useMap = false")
	}
	s.steps = append(s.steps, v)
}

func (s *store) previous() map[string]interface{} {
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

func (s *store) latest() map[string]interface{} {
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

func (s *store) toNormalizedMap() map[string]interface{} {
	store := map[string]interface{}{}
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

func (s *store) toMap() map[string]interface{} {
	store := map[string]interface{}{}
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
