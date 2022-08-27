package runn

const (
	storeVarsKey     = "vars"
	storeStepsKey    = "steps"
	storeParentKey   = "parent"
	storeIncludedKey = "included"
	storeFuncValue   = "[func]"
)

type store struct {
	steps      []map[string]interface{}
	stepMap    map[string]map[string]interface{}
	vars       map[string]interface{}
	funcs      map[string]interface{}
	bindVars   map[string]interface{}
	parentVars map[string]interface{}
	useMap     bool // Use map syntax in `steps:`.
	loopIndex  *int
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
