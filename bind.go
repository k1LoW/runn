package runn

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/antonmedv/expr/ast"
	"github.com/antonmedv/expr/parser"
	"github.com/samber/lo"
)

const bindRunnerKey = "bind"

type bindRunner struct{}

func newBindRunner() *bindRunner {
	return &bindRunner{}
}

func (rnr *bindRunner) Run(ctx context.Context, s *step, first bool) error {
	o := s.parent
	cond := s.bindCond
	store := o.store.toMap()
	store[storeRootKeyIncluded] = o.included
	if first {
		store[storeRootPrevious] = o.store.latest()
	} else {
		store[storeRootPrevious] = o.store.previous()
		store[storeRootKeyCurrent] = o.store.latest()
	}
	keys := lo.Keys(cond)
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if lo.Contains(reservedStoreRootKeys, k) {
			return fmt.Errorf("%q is reserved", k)
		}
		v := cond[k]
		kv, err := evalBindKeyValue(o.store.bindVars, k, v, store)
		if err != nil {
			return err
		}
		o.store.bindVars = kv
	}
	if first {
		o.record(nil)
	}
	return nil
}

func evalBindKeyValue(bindVars map[string]any, k string, v any, store map[string]any) (map[string]any, error) {
	vv, err := EvalAny(v, store)
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
		if lo.Contains(reservedStoreRootKeys, k) {
			return nil, fmt.Errorf("%q is reserved", k)
		}
		m[k] = v
	case *ast.MemberNode:
		switch nnn := nn.Node.(type) {
		case *ast.IdentifierNode:
			k := nnn.Value
			if lo.Contains(reservedStoreRootKeys, k) {
				return nil, fmt.Errorf("%q is reserved", k)
			}
			switch p := nn.Property.(type) {
			case *ast.IdentifierNode:
				kk, err := EvalAny(p.Value, store)
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
				kk, err := EvalAny(p.String(), store)
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
				kk, err := EvalAny(p.Value, store)
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
				kk, err := EvalAny(p.String(), store)
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

func mergeVars(store map[string]any, vars map[string]any) map[string]any {
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
				svv2 := make(map[any]any)
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
				vv2 := make(map[any]any)
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

func mergeMapAny(store map[any]any, vars map[any]any) map[any]any {
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
