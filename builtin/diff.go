package builtin

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/itchyny/gojq"
)

func Diff(x, y any, ignores ...any) (string, error) {
	var ignoreSpecifiers []string
	for _, i := range ignores {
		switch v := i.(type) {
		case string:
			ignoreSpecifiers = append(ignoreSpecifiers, v)
		case []any:
			for _, vv := range v {
				s, ok := vv.(string)
				if !ok {
					return "", fmt.Errorf("invalid ignore specifiers: %v", vv)
				}
				ignoreSpecifiers = append(ignoreSpecifiers, s)
			}
		case []string:
			for _, s := range v {
				ignoreSpecifiers = append(ignoreSpecifiers, s)
			}
		default:
			return "", fmt.Errorf("invalid ignore specifiers: %v", i)
		}
	}

	impl := diffImpl{}

	// normalize values
	vx, err := impl.normalizeInput(x)
	if err != nil {
		return "", err
	}
	vy, err := impl.normalizeInput(y)
	if err != nil {
		return "", err
	}

	jqIgnorePaths, cmpIgnoreKeys := impl.splitIgnoreSpecifiers(ignoreSpecifiers)

	var diffOpts []cmp.Option
	if len(cmpIgnoreKeys) >= 0 {
		diffOpts = append(diffOpts, impl.buildFilterForCmpIgnoreKeys(cmpIgnoreKeys))
	}

	if len(jqIgnorePaths) > 0 {
		if filter, err := impl.buildFilterForJqIgnorePaths(jqIgnorePaths, vx, vy); err != nil {
			return "", err
		} else {
			diffOpts = append(diffOpts, filter)
		}
	}

	return cmp.Diff(vx, vy, diffOpts...), nil
}

type diffImpl struct{}

func (d *diffImpl) normalizeInput(x any) (any, error) {
	bx, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}

	var vx any
	err = json.Unmarshal(bx, &vx)
	if err != nil {
		return nil, err
	}

	return vx, nil
}

func (d *diffImpl) buildExpandPathsJqQuery(pathExpressions []string) (*gojq.Query, error) {
	qb := strings.Builder{}
	qb.WriteString("[")
	for i, pathExpr := range pathExpressions {
		if i > 0 {
			qb.WriteString(", ")
		}
		qb.WriteString("(try path(")
		qb.WriteString(pathExpr)
		qb.WriteString("))")
	}
	qb.WriteString("]")

	query, err := gojq.Parse(qb.String())
	if err != nil {
		return nil, err
	}

	return query, nil
}

func (d *diffImpl) buildIgnoreTransformJqQuery(ignorePaths []string) (*gojq.Query, error) {
	qb := strings.Builder{}
	qb.WriteString("delpaths([")
	for i, pathExpr := range ignorePaths {
		if i > 0 {
			qb.WriteString(", ")
		}
		qb.WriteString("(try path(")
		qb.WriteString(pathExpr)
		qb.WriteString("))")
	}
	qb.WriteString("])")

	query, err := gojq.Parse(qb.String())
	if err != nil {
		return nil, fmt.Errorf("failed to build the ignorePaths query: %w", err)
	}

	return query, nil
}

func (d *diffImpl) splitIgnoreSpecifiers(ignoreSpecifiers []string) ([]string, []string) {
	jqIgnorePaths := make([]string, 0)
	cmpIgnoreKeys := make([]string, 0)

	for _, specifier := range ignoreSpecifiers {
		if strings.HasPrefix(specifier, ".") {
			// jq path syntax
			jqIgnorePaths = append(jqIgnorePaths, specifier)
		} else {
			// by map key string for backward compatibility
			cmpIgnoreKeys = append(cmpIgnoreKeys, specifier)
		}
	}

	return jqIgnorePaths, cmpIgnoreKeys
}

func (d *diffImpl) buildFilterForCmpIgnoreKeys(cmpIgnoreKeys []string) cmp.Option {
	cmpIgnoreKeySet := make(map[string]struct{}, len(cmpIgnoreKeys))
	for _, key := range cmpIgnoreKeys {
		cmpIgnoreKeySet[key] = struct{}{}
	}

	return cmpopts.IgnoreMapEntries(func(key string, val any) bool {
		_, ignored := cmpIgnoreKeySet[key]
		return ignored
	})
}

func (d *diffImpl) buildFilterForJqIgnorePaths(jqIgnorePaths []string, vx any, vy any) (cmp.Option, error) {
	query, err := d.buildExpandPathsJqQuery(jqIgnorePaths)
	if err != nil {
		return nil, fmt.Errorf("diff ignorePaths query parsing error: %w", err)
	}

	code, err := gojq.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("diff ignorePaths query compile error: %w", err)
	}

	pathBuilder := newJqPathBuilder()
	pathLookup := newJqPathLookup()

	if err := d.registerToJqPathLookup(pathLookup, pathBuilder, code, vx); err != nil {
		return nil, err
	}

	if err := d.registerToJqPathLookup(pathLookup, pathBuilder, code, vy); err != nil {
		return nil, err
	}

	filterFunc := func(p cmp.Path) bool {
		px, err := pathBuilder.fromCmpPath(p, false)
		if err == nil && px != nil && pathLookup.isExist(px) {
			return true
		}

		py, err := pathBuilder.fromCmpPath(p, true)
		if err == nil && py != nil && pathLookup.isExist(py) {
			return true
		}

		return false
	}

	return cmp.FilterPath(filterFunc, cmp.Ignore()), nil
}

func (d *diffImpl) registerToJqPathLookup(pathLookup *jqPathLookup, pathBuilder *jqPathBuilder, pathExpressionsCode *gojq.Code, input any) error {
	if paths, err := d.applyJqQueryCompiled(pathExpressionsCode, input); err != nil {
		return fmt.Errorf("applying diff ignorePaths error: %w", err)
	} else {
		for _, path := range paths.([]any) {
			if p2, err := pathBuilder.fromAnyArray(path.([]any)); err != nil {
				return err
			} else {
				pathLookup.put(p2)
			}
		}
	}
	return nil
}

func (d *diffImpl) applyJqQueryCompiled(code *gojq.Code, input any) (any, error) {
	iter := code.Run(input)
	for {
		out, ok := iter.Next()
		if !ok {
			break
		}

		if err, ok := out.(error); ok {
			var haltErr *gojq.HaltError
			if errors.As(err, &haltErr) && haltErr.Value() == nil {
				break
			}

			return nil, err
		}

		return out, nil
	}
	return input, nil
}

type jqPathEntryType int8

const (
	jqPathEntryRoot   = -1
	jqPathEntryString = 1
	jqPathEntryInt    = 2
)

type jqPathEntry struct {
	t jqPathEntryType
	s string
	i int
}

func (e jqPathEntry) String() string { //nostyle:recvtype
	switch e.t {
	case jqPathEntryRoot:
		return ""
	case jqPathEntryString:
		return e.s
	case jqPathEntryInt:
		return fmt.Sprintf("[%d]", e.i)
	default:
		return "?"
	}
}

type jqPath []jqPathEntry

func (p jqPath) String() string { //nostyle:recvtype
	sb := strings.Builder{}
	for _, entry := range p {
		sb.WriteRune('.')
		sb.WriteString(entry.String())
	}
	return sb.String()
}

type jqPathTreeNode struct {
	pathEntry    jqPathEntry
	children     map[jqPathEntry]*jqPathTreeNode
	intermediate bool
}

type jqPathLookup struct {
	pathTreeRoot *jqPathTreeNode
}

func newJqPathLookup() *jqPathLookup {
	return &jqPathLookup{
		pathTreeRoot: &jqPathTreeNode{pathEntry: jqPathEntry{t: jqPathEntryRoot}, intermediate: true},
	}
}

func (l *jqPathLookup) put(path jqPath) bool {
	node := l.pathTreeRoot

	for _, entry := range path {
		if node.children == nil {
			node.children = make(map[jqPathEntry]*jqPathTreeNode)
		}

		if child, exists := node.children[entry]; exists {
			node = child
		} else {
			newChild := &jqPathTreeNode{pathEntry: entry, intermediate: true}
			node.children[entry] = newChild
			node = newChild
		}
	}

	added := false
	if node.intermediate {
		node.intermediate = false
		added = true
	}

	return added
}

func (l *jqPathLookup) isExist(path jqPath) bool {
	node := l.pathTreeRoot

	for i, entry := range path {
		isRoot := i == 0

		if isRoot && entry.t == jqPathEntryRoot {
			continue
		} else {
			if node.children == nil {
				return false
			}

			if child, exists := node.children[entry]; exists {
				node = child
			} else {
				return false
			}
		}
	}

	return !node.intermediate
}

type jqPathBuilder struct {
	work []jqPathEntry
}

func newJqPathBuilder() *jqPathBuilder {
	return &jqPathBuilder{work: make([]jqPathEntry, 0)}
}

func (b *jqPathBuilder) reset() {
	b.work = b.work[:0]
}

func (b *jqPathBuilder) add(entry jqPathEntry) error {
	b.work = append(b.work, entry)
	return nil
}

func (b *jqPathBuilder) addAny(entry any) error {
	switch v := entry.(type) {
	case string:
		return b.add(jqPathEntry{t: jqPathEntryString, s: v})
	case int:
		return b.add(jqPathEntry{t: jqPathEntryInt, i: v})
	default:
		return fmt.Errorf("type %T cannot be used for jq path", v)
	}
}

func (b *jqPathBuilder) addCmpPathStep(ps cmp.PathStep, isRoot bool, xySel bool) error {
	if isRoot {
		return b.add(jqPathEntry{t: jqPathEntryRoot})
	}

	if mi, ok := ps.(cmp.MapIndex); ok {
		if mi.Key().Kind() == reflect.String {
			return b.add(jqPathEntry{t: jqPathEntryString, s: mi.Key().String()})
		} else {
			return fmt.Errorf("non-string map key is not expected")
		}
	}

	if si, ok := ps.(cmp.SliceIndex); ok {
		ix, iy := si.SplitKeys()
		if !xySel {
			return b.add(jqPathEntry{t: jqPathEntryInt, i: ix})
		} else {
			return b.add(jqPathEntry{t: jqPathEntryInt, i: iy})
		}
	}

	return nil
}

func (b *jqPathBuilder) build() jqPath {
	return b.work
}

func (b *jqPathBuilder) fromAnyArray(p []any) (jqPath, error) {
	b.reset()
	for _, a := range p {
		err := b.addAny(a)
		if err != nil {
			return nil, err
		}
	}
	return b.build(), nil
}

func (b *jqPathBuilder) fromCmpPath(p cmp.Path, sel bool) (jqPath, error) {
	b.reset()
	for i, ps := range p {
		err := b.addCmpPathStep(ps, i == 0, sel)
		if err != nil {
			return nil, err
		}
	}
	return b.build(), nil
}
