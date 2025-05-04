package store

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/goccy/go-json"
	"github.com/k1LoW/maskedio"
	"github.com/k1LoW/runn/internal/expr"
	"github.com/k1LoW/runn/internal/kv"
	"github.com/mattn/go-isatty"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

const (
	RootKeyVars           = "vars"
	RootKeySteps          = "steps"
	RootKeyParent         = "parent"
	RootKeyIncluded       = "included"
	RootKeyCurrent        = "current"
	RootKeyPrevious       = "previous"
	RootKeyEnv            = "env"
	RootKeyCookie         = "cookies"
	RootKeyNodes          = "nodes"
	RootKeyParams         = "params"
	RootKeyRunn           = "runn"
	RootKeyNeeds          = "needs"
	RootKeyLoopCountIndex = "i"
)

const (
	StepKeyOutcome = "outcome"
	FuncValue      = "[func]"
)

// `runn` is global store for runn.
const (
	RunnKeyKV        = "kv"
	RunnKeyRunNIndex = "i"
	RunnKeyStdin     = "stdin"
)

var stdin any

// Reserved store root keys.
var ReservedRootKeys = []string{
	RootKeyVars,
	RootKeySteps,
	RootKeyParent,
	RootKeyIncluded,
	RootKeyCurrent,
	RootKeyPrevious,
	RootKeyEnv,
	RootKeyCookie,
	RootKeyNodes,
	RootKeyParams,
	RootKeyLoopCountIndex,
	RootKeyRunn,
	RootKeyNeeds,
}

type Store struct {
	stepList   map[int]map[string]any
	stepKeys   []string // Keys for `steps:` in map syntax.
	vars       map[string]any
	funcs      map[string]any
	bindVars   map[string]any
	parentVars map[string]any
	needsVars  map[string]any
	useMap     bool // Use map syntax in `steps:`.
	loopIndex  *int
	cookies    map[string]map[string]*http.Cookie
	kv         *kv.KV
	runNIndex  int

	// for secret masking
	secrets []string // Secret var names to be masked.
	mr      *maskedio.Rule
}

func New(vars, funcs map[string]any, secrets []string, stepKeys []string) *Store {
	useMap := len(stepKeys) > 0
	s := &Store{
		stepList:   map[int]map[string]any{},
		stepKeys:   stepKeys,
		vars:       vars,
		funcs:      funcs,
		bindVars:   map[string]any{},
		parentVars: map[string]any{},
		needsVars:  map[string]any{},
		useMap:     useMap,
		runNIndex:  -1,
		secrets:    secrets,
		mr:         maskedio.NewRule(),
	}
	s.SetMaskKeywords(s.ToMap())

	return s
}

func (s *Store) Record(idx int, v map[string]any) {
	s.stepList[idx] = v
}

func (s *Store) Cookies() map[string]map[string]*http.Cookie {
	return s.cookies
}

func (s *Store) StepKeys() []string {
	return s.stepKeys
}

func (s *Store) SetVar(k string, v any) {
	s.vars[k] = v
}

func (s *Store) SetNeedsVar(k string, ns *Store) {
	s.needsVars[k] = ns.bindVars
}

func (s *Store) SetBindVar(k string, v any) error {
	if lo.Contains(ReservedRootKeys, k) {
		return fmt.Errorf("%q is reserved", k)
	}
	s.bindVars[k] = v
	return nil
}

func (s *Store) RecordBindVar(k string, v any, sm map[string]any) error {
	if lo.Contains(ReservedRootKeys, k) {
		return fmt.Errorf("%q is reserved", k)
	}
	kv, err := evalBindKeyValue(s.bindVars, k, v, sm)
	if err != nil {
		return err
	}
	s.bindVars = kv
	return nil
}

func (s *Store) SetParentVars(vars map[string]any) {
	s.parentVars = vars
}

func (s *Store) KV() *kv.KV {
	return s.kv
}

func (s *Store) SetKV(kv *kv.KV) {
	s.kv = kv
}

func (s *Store) SetSecrets(secrets []string) {
	s.secrets = secrets
}

func (s *Store) SetRunNIndex(i int) {
	s.runNIndex = i
}

func (s *Store) LoopIndex() *int {
	return s.loopIndex
}

func (s *Store) SetLoopIndex(i int) {
	s.loopIndex = &i
}

func (s *Store) ClearLoopIndex() {
	s.loopIndex = nil
}

func (s *Store) RunNIndex() int {
	return s.runNIndex
}

func (s *Store) Funcs() map[string]any {
	return s.funcs
}

func (s *Store) StepLen() int {
	return len(s.stepList)
}

func (s *Store) Previous() map[string]any {
	if s.StepLen() < 2 {
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
	for i := range s.stepList {
		for j := range s.stepKeys {
			if i == j {
				idxs = append(idxs, i)
			}
		}
	}
	slices.Sort(idxs)
	if v, ok := s.stepList[idxs[len(idxs)-2]]; ok {
		return v
	}
	return nil
}

func (s *Store) Latest() map[string]any {
	if s.StepLen() == 0 {
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
	for i := range s.stepList {
		for j := range s.stepKeys {
			if i == j {
				idxs = append(idxs, i)
			}
		}
	}
	slices.Sort(idxs)
	if v, ok := s.stepList[idxs[len(idxs)-1]]; ok {
		return v
	}
	return nil
}

func (s *Store) RecordTo(idx int, key string, value any) error {
	if s.StepLen() == 0 {
		return errors.New("failed to record: store.steps is zero")
	}
	if _, ok := s.stepList[idx]; ok {
		s.stepList[idx][key] = value
		return nil
	}
	return errors.New("failed to record")
}

func (s *Store) RecordCookie(cookies []*http.Cookie) {
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

func (s *Store) ToMap() map[string]any {
	store := map[string]any{}
	store[RootKeyEnv] = envMap()
	maps.Copy(store, s.funcs)
	store[RootKeyVars] = s.vars
	if s.useMap {
		store[RootKeySteps] = convertStepListToMap(s.stepList, s.stepKeys)
	} else {
		store[RootKeySteps] = convertStepListToList(s.stepList)
	}
	if len(s.parentVars) > 0 {
		store[RootKeyParent] = s.parentVars
	} else {
		store[RootKeyParent] = nil
	}
	if len(s.needsVars) > 0 {
		store[RootKeyNeeds] = s.needsVars
	} else {
		store[RootKeyNeeds] = nil
	}
	maps.Copy(store, s.bindVars)
	if s.loopIndex != nil {
		store[RootKeyLoopCountIndex] = *s.loopIndex
	}
	if s.cookies != nil {
		store[RootKeyCookie] = s.cookies
	}

	runnm := map[string]any{}
	// runn.kv
	if s.kv != nil {
		kv := map[string]any{}
		for _, k := range s.kv.Keys() {
			kv[k] = s.kv.Get(k)
		}
		runnm[RunnKeyKV] = kv
	}
	// runn.i
	runnm[RunnKeyRunNIndex] = s.runNIndex
	// runn.stdin
	if stdin != nil {
		runnm[RunnKeyStdin] = stdin
	}
	store[RootKeyRunn] = runnm

	s.SetMaskKeywords(store)

	return store
}

// ToMapForIncludeRunner - returns a map for include runner.
// toMap without s.parentVars s.needsVars and runn.* .
func (s *Store) ToMapForIncludeRunner() map[string]any {
	store := map[string]any{}
	store[RootKeyEnv] = envMap()
	for k := range s.funcs {
		store[k] = FuncValue
	}
	store[RootKeyVars] = s.vars
	if s.useMap {
		store[RootKeySteps] = convertStepListToMap(s.stepList, s.stepKeys)
	} else {
		store[RootKeySteps] = convertStepListToList(s.stepList)
	}
	maps.Copy(store, s.bindVars)
	if s.loopIndex != nil {
		store[RootKeyLoopCountIndex] = *s.loopIndex
	}
	if s.cookies != nil {
		store[RootKeyCookie] = s.cookies
	}
	s.SetMaskKeywords(store)

	return store
}

// ToMapForDbg - returns a map for dbg.
// toMap without s.funcs.
func (s *Store) ToMapForDbg() map[string]any {
	store := map[string]any{}
	store[RootKeyEnv] = envMap()
	store[RootKeyVars] = s.vars
	if s.useMap {
		store[RootKeySteps] = convertStepListToMap(s.stepList, s.stepKeys)
	} else {
		store[RootKeySteps] = convertStepListToList(s.stepList)
	}
	if len(s.parentVars) > 0 {
		store[RootKeyParent] = s.parentVars
	} else {
		store[RootKeyParent] = nil
	}
	if len(s.needsVars) > 0 {
		store[RootKeyNeeds] = s.needsVars
	} else {
		store[RootKeyNeeds] = nil
	}
	maps.Copy(store, s.bindVars)
	if s.loopIndex != nil {
		store[RootKeyLoopCountIndex] = *s.loopIndex
	}
	if s.cookies != nil {
		store[RootKeyCookie] = s.cookies
	}

	runnm := map[string]any{}
	// runn.kv
	if s.kv != nil {
		kv := map[string]any{}
		for _, k := range s.kv.Keys() {
			kv[k] = s.kv.Get(k)
		}
		runnm[RunnKeyKV] = kv
	}
	// runn.i
	runnm[RunnKeyRunNIndex] = s.runNIndex
	// runn.stdin
	if stdin != nil {
		runnm[RunnKeyStdin] = stdin
	}
	store[RootKeyRunn] = runnm

	// runn.stdin
	if stdin != nil {
		store[RunnKeyStdin] = stdin
	}

	s.SetMaskKeywords(store)

	return store
}

func (s *Store) SetMaskKeywords(store map[string]any) {
	store[RootKeyCurrent] = s.Latest()
	defer func() {
		delete(store, RootKeyCurrent)
	}()

	for _, key := range s.secrets {
		v, err := expr.Eval(key, store)
		if err != nil {
			continue
		}

		switch vv := v.(type) {
		case map[string]any:
		case []any:
		default:
			s.MaskRule().SetKeyword(cast.ToString(vv))
		}
	}
}

func (s *Store) ClearSteps() {
	s.stepList = map[int]map[string]any{}
	// keep stepKeys, vars, bindVars, cookies, kv, parentVars, runNIndex

	s.loopIndex = nil
}

func (s *Store) SetMaskRule(mr *maskedio.Rule) {
	s.mr = mr
}

func (s *Store) MaskRule() *maskedio.Rule {
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

func convertStepListToMap(l map[int]map[string]any, keys []string) map[string]map[string]any {
	m := map[string]map[string]any{}
	for i, v := range l {
		key := keys[i]
		m[key] = v
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

func evalBindKeyValue(bindVars map[string]any, k string, v any, store map[string]any) (map[string]any, error) {
	vv, err := expr.EvalAny(v, store)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(k, "[]") {
		// Append to slice
		// - foo[]
		// - foo[bar][]
		kk := strings.TrimSuffix(k, "[]")
		return evalBindKeyValue(bindVars, kk, []any{v}, store)
	}
	// Merge to map
	// - foo
	// - foo[bar]
	// - foo['bar']
	// - foo[5]
	// - foo[bar][baz]
	tr, err := parser.Parse(k)
	if err != nil {
		return nil, err
	}
	kv, err := nodeToMap(tr.Node, vv, store)
	if err != nil {
		return nil, err
	}
	return mergeVars(bindVars, kv), nil
}

func nodeToMap(n ast.Node, v any, store map[string]any) (map[string]any, error) {
	m := map[string]any{}
	switch nn := n.(type) {
	case *ast.IdentifierNode:
		k := nn.Value
		if lo.Contains(ReservedRootKeys, k) {
			return nil, fmt.Errorf("%q is reserved", k)
		}
		m[k] = v
	case *ast.MemberNode:
		switch nnn := nn.Node.(type) {
		case *ast.IdentifierNode:
			k := nnn.Value
			if lo.Contains(ReservedRootKeys, k) {
				return nil, fmt.Errorf("%q is reserved", k)
			}
			switch p := nn.Property.(type) {
			case *ast.IdentifierNode:
				kk, err := expr.EvalAny(p.Value, store)
				if err != nil {
					return nil, err
				}
				if kk == nil {
					return nil, fmt.Errorf("invalid value: %v", p.Value)
				}
				m[k] = map[any]any{
					kk: v,
				}
			case *ast.StringNode:
				m[k] = map[any]any{
					p.Value: v,
				}
			case *ast.IntegerNode:
				m[k] = map[any]any{
					p.Value: v,
				}
			case *ast.MemberNode:
				kk, err := expr.EvalAny(p.String(), store)
				if err != nil {
					return nil, err
				}
				if kk == nil {
					return nil, fmt.Errorf("invalid value: %v", p.String())
				}
				m[k] = map[any]any{
					kk: v,
				}
			default:
				return nil, fmt.Errorf("invalid node type of %v: %T", nn.Property, nn.Property)
			}
		case *ast.MemberNode:
			var vv map[any]any
			switch p := nn.Property.(type) {
			case *ast.IdentifierNode:
				kk, err := expr.EvalAny(p.Value, store)
				if err != nil {
					return nil, err
				}
				if kk == nil {
					return nil, fmt.Errorf("invalid value: %v", p.Value)
				}
				vv = map[any]any{
					kk: v,
				}
			case *ast.StringNode:
				vv = map[any]any{
					p.Value: v,
				}
			case *ast.IntegerNode:
				vv = map[any]any{
					p.Value: v,
				}
			case *ast.MemberNode:
				kk, err := expr.EvalAny(p.String(), store)
				if err != nil {
					return nil, err
				}
				if kk == nil {
					return nil, fmt.Errorf("invalid value: %v", p.String())
				}
				vv = map[any]any{
					kk: v,
				}
			default:
				return nil, fmt.Errorf("invalid node type of %v: %T", nn.Property, nn.Property)
			}
			vvv, err := nodeToMap(nnn, vv, store)
			if err != nil {
				return nil, err
			}
			m = vvv
		}
	default:
		return nil, fmt.Errorf("invalid node type of %v: %T", n, n)
	}
	return m, nil
}

func mergeVars(org map[string]any, vars map[string]any) map[string]any {
	store := make(map[string]any, len(org)+len(vars))
	maps.Copy(store, org)
	for k, v := range vars {
		sv, ok := store[k]
		if !ok {
			store[k] = v
			continue
		}
		switch svv := sv.(type) {
		case map[string]any:
			switch vv := v.(type) {
			case map[string]any:
				store[k] = mergeVars(svv, vv)
			case map[any]any:
				// convert svv map[string]any to map[any]any
				svv2 := make(map[any]any, len(svv))
				for k, v := range svv {
					svv2[k] = v
				}
				store[k] = mergeMapAny(svv2, vv)
			default:
				store[k] = vv
			}
		case map[any]any:
			switch vv := v.(type) {
			case map[string]any:
				// convert vv map[string]any to map[any]any
				vv2 := make(map[any]any, len(vv))
				for k, v := range vv {
					vv2[k] = v
				}
				store[k] = mergeMapAny(svv, vv2)
			case map[any]any:
				store[k] = mergeMapAny(svv, vv)
			default:
				store[k] = vv
			}
		case []any:
			switch vv := v.(type) {
			case []any:
				store[k] = append(svv, vv...)
			default:
				store[k] = vv
			}
		default:
			store[k] = v
		}
	}
	return store
}

func mergeMapAny(org map[any]any, vars map[any]any) map[any]any {
	store := map[any]any{}
	maps.Copy(store, org)
	for k, v := range vars {
		sv, ok := store[k]
		if !ok {
			store[k] = v
			continue
		}
		switch svv := sv.(type) {
		case map[string]any:
			switch vv := v.(type) {
			case map[string]any:
				store[k] = mergeVars(svv, vv)
			default:
				store[k] = vv
			}
		case map[any]any:
			switch vv := v.(type) {
			case map[any]any:
				store[k] = mergeMapAny(svv, vv)
			default:
				store[k] = vv
			}
		case []any:
			switch vv := v.(type) {
			case []any:
				store[k] = append(svv, vv...)
			default:
				store[k] = vv
			}
		default:
			store[k] = v
		}
	}
	return store
}
