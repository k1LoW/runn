package exprtrace

import (
	"context"
	"maps"
	"reflect"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	exprbuiltin "github.com/expr-lang/expr/builtin"
)

type Tracer struct {
	trace          *EvalTraceStore
	store          EvalEnv
	contextType    reflect.Type
	traceTagType   reflect.Type
	funcAttrsCache map[string]int
	builtinsMap    map[string]*exprbuiltin.Function
	eval           patcherEvaluationPhaseFields

	firstPhasePatcher  firstPhasePatcher
	secondPhasePatcher secondPhasePatcher
}

type firstPhasePatcher struct {
	trace                *EvalTraceStore
	mapper               *TagMapper
	structFieldBaseNodes map[uintptr]bool // Base nodes of struct field accesses that should not be patched
}

type secondPhasePatcher struct {
	patching             patcherPatchingPhaseFields
	mapper               *TagMapper
	structFieldBaseNodes *map[uintptr]bool // Pointer to firstPhasePatcher's structFieldBaseNodes
}

type patcherPatchingPhaseFields struct {
	closurePointerNodes   []*PointerNodeTraceInfo
	builtinPredicateNodes []*PredicateNodeTraceInfo
}

type patcherEvaluationPhaseFields struct {
	closureEvalCount map[TraceStoreKey]int
}

const (
	keyTracerFuncInteger           = "tracer.integer"
	keyTracerFuncFloat             = "tracer.float"
	keyTracerFuncCall              = "tracer.call"
	keyTracerFuncPredicate         = "tracer.predicate"
	keyTracerFuncBuiltin           = "tracer.builtin"
	keyTracerFuncBinary            = "tracer.binary"
	keyTracerFuncConditional       = "tracer.conditional"
	keyTracerFuncIdentifier        = "tracer.identifier"
	keyTracerFuncPointer           = "tracer.pointer"
	keyTracerFuncVariableDecorator = "tracer.variable_declarator"
	keyTracerFuncMember            = "tracer.member"
	keyTracerFuncArray             = "tracer.array"
	keyTracerFuncSlice             = "tracer.slice"
	keyTracerFuncMap               = "tracer.map"
	keyTracerFuncPairValue         = "tracer.pair_value"
)

var (
	identifierTracerFuncInteger           = ast.IdentifierNode{Value: keyTracerFuncInteger}
	identifierTracerFuncFloat             = ast.IdentifierNode{Value: keyTracerFuncFloat}
	identifierTracerFuncCall              = ast.IdentifierNode{Value: keyTracerFuncCall}
	identifierTracerFuncPredicate         = ast.IdentifierNode{Value: keyTracerFuncPredicate}
	identifierTracerFuncBuiltin           = ast.IdentifierNode{Value: keyTracerFuncBuiltin}
	identifierTracerFuncBinary            = ast.IdentifierNode{Value: keyTracerFuncBinary}
	identifierTracerFuncConditional       = ast.IdentifierNode{Value: keyTracerFuncConditional}
	identifierTracerFuncIdentifier        = ast.IdentifierNode{Value: keyTracerFuncIdentifier}
	identifierTracerFuncPointer           = ast.IdentifierNode{Value: keyTracerFuncPointer}
	identifierTracerFuncVariableDecorator = ast.IdentifierNode{Value: keyTracerFuncVariableDecorator}
	identifierTracerFuncMember            = ast.IdentifierNode{Value: keyTracerFuncMember}
	identifierTracerFuncArray             = ast.IdentifierNode{Value: keyTracerFuncArray}
	identifierTracerFuncSlice             = ast.IdentifierNode{Value: keyTracerFuncSlice}
	identifierTracerFuncMap               = ast.IdentifierNode{Value: keyTracerFuncMap}
	identifierTracerFuncPairValue         = ast.IdentifierNode{Value: keyTracerFuncPairValue}
)

func (t *Tracer) InstallTracerFunctions(store any) any {
	var env map[string]any

	// Handle both map[string]any and EvalEnv types
	switch s := store.(type) {
	case map[string]any:
		env = maps.Clone(s)
	case EvalEnv:
		env = maps.Clone(map[string]any(s))
	default:
		// If it's neither, create a new empty map
		env = make(map[string]any)
	}

	env[keyTracerFuncInteger] = t.traceInteger
	env[keyTracerFuncFloat] = t.traceFloat
	env[keyTracerFuncCall] = t.traceCall
	env[keyTracerFuncPredicate] = t.tracePredicate
	env[keyTracerFuncBuiltin] = t.traceBuiltin
	env[keyTracerFuncBinary] = t.traceBinary
	env[keyTracerFuncConditional] = t.traceConditional
	env[keyTracerFuncIdentifier] = t.traceIdentifier
	env[keyTracerFuncPointer] = t.tracePointer
	env[keyTracerFuncVariableDecorator] = t.traceVariableDeclarator
	env[keyTracerFuncMember] = t.traceMember
	env[keyTracerFuncArray] = t.traceArray
	env[keyTracerFuncSlice] = t.traceSlice
	env[keyTracerFuncMap] = t.traceMap
	env[keyTracerFuncPairValue] = t.tracePairValue
	return env
}

type NodeEvalResult interface { //nostyle:ifacenames
	Output() any
}

type TraceEntry interface { //nostyle:ifacenames
	EvalResultByCallCount(callCount int) NodeEvalResult
}

type traceEntry[T NodeEvalResult] struct {
	tag         EvalTraceTag
	evalResults []T
}

func (t *traceEntry[T]) EvalResultByCallCount(callCount int) NodeEvalResult {
	return t.evalResults[callCount]
}

type callEvalResult struct {
	output any
}

func (c *callEvalResult) Output() any {
	return c.output
}

type closureEvalResult struct {
	output any
}

func (c *closureEvalResult) Output() any {
	return c.output
}

type builtinEvalResult struct {
	output           any
	closureEvalCount int
}

func (b *builtinEvalResult) Output() any {
	return b.output
}

type binaryEvalResult struct {
	output any
}

func (b *binaryEvalResult) Output() any {
	return b.output
}

type conditionalEvalResult struct {
	output any
}

func (c *conditionalEvalResult) Output() any {
	return c.output
}

type identifierEvalResult struct {
	value any
}

func (i *identifierEvalResult) Output() any {
	return i.value
}

type pointerEvalResult struct {
	value            any
	closureCallCount int
}

func (p *pointerEvalResult) Output() any {
	return p.value
}

type variableDeclaratorEvalResult struct {
	value any
}

func (v *variableDeclaratorEvalResult) Output() any {
	return v.value
}

type memberEvalResult struct {
	value    any
	optional bool
	method   bool
}

func (m *memberEvalResult) Output() any {
	return m.value
}

type mapEvalResult struct {
	value map[string]any
}

func (m *mapEvalResult) Output() any {
	return m.value
}

type boolEvalResult struct {
	value any
}

func (b *boolEvalResult) Output() any {
	return b.value
}

type integerEvalResult struct {
	value any
}

func (i *integerEvalResult) Output() any {
	return i.value
}

type floatEvalResult struct {
	value any
}

func (f *floatEvalResult) Output() any {
	return f.value
}

type pairEvalResult struct {
	value any
}

func (p *pairEvalResult) Output() any {
	return p.value
}

type arrayEvalResult struct {
	values any
}

func (a *arrayEvalResult) Output() any {
	return a.values
}

type sliceEvalResult struct {
	values any
}

func (a *sliceEvalResult) Output() any {
	return a.values
}

type baseTraceInfo struct {
	tag EvalTraceTag
}

type IntegerNodeTraceInfo struct {
	baseTraceInfo
	integerNode *ast.IntegerNode
}

type FloatNodeTraceInfo struct {
	baseTraceInfo
	floatNode *ast.FloatNode
}

type CallNodeTraceInfo struct {
	baseTraceInfo
	callNode *ast.CallNode
}

type PredicateNodeTraceInfo struct {
	baseTraceInfo
	predicateNode *ast.PredicateNode
	builtinTag    EvalTraceTag
}

type BuiltinNodeTraceInfo struct {
	baseTraceInfo
	builtinNode *ast.BuiltinNode
}

type BinaryNodeTraceInfo struct {
	baseTraceInfo
	binaryNode *ast.BinaryNode
}

type ArrayNodeTraceInfo struct {
	baseTraceInfo
	arrayNode *ast.ArrayNode
}

type SliceNodeTraceInfo struct {
	baseTraceInfo
	sliceNode *ast.SliceNode
}

type MapNodeTraceInfo struct {
	baseTraceInfo
	mapNode *ast.MapNode
}

type PairNodeTraceInfo struct {
	baseTraceInfo
	pairNode *ast.PairNode
}

type ConditionalNodeTraceInfo struct {
	baseTraceInfo
	conditionalNode *ast.ConditionalNode
}

type IdentifierNodeTraceInfo struct {
	baseTraceInfo
	identifierNode *ast.IdentifierNode
}

type PointerNodeTraceInfo struct {
	baseTraceInfo
	pointerNode *ast.PointerNode
	closureTag  EvalTraceTag
}

type VariableDeclaratorNodeTraceInfo struct {
	baseTraceInfo
	variableDeclaratorNode *ast.VariableDeclaratorNode
}

type MemberNodeTraceInfo struct {
	baseTraceInfo
	memberNode *ast.MemberNode
}

type TagMapper struct {
	ptrMap  map[uintptr]int
	counter int
}

func (m *TagMapper) build(node ast.Node) {
	ast.Walk(&node, m)
}

func (m *TagMapper) Visit(node *ast.Node) {
	cnt := m.counter
	cnt += 1
	m.counter = cnt

	ptr := reflect.ValueOf(*node).Pointer()
	m.ptrMap[ptr] = cnt
}

func (m *TagMapper) indexByNode(node ast.Node) int {
	ptr := reflect.ValueOf(node).Pointer()
	if idx, ok := m.ptrMap[ptr]; ok {
		return idx
	} else {
		return -1
	}
}

func NewTracer(trace *EvalTraceStore, store EvalEnv) *Tracer {
	builtinFunctionsMap := map[string]*exprbuiltin.Function{}

	for _, function := range exprbuiltin.Builtins {
		builtinFunctionsMap[function.Name] = function
	}
	mapper := &TagMapper{ptrMap: map[uintptr]int{}, counter: 0}
	structFieldBaseNodes := map[uintptr]bool{}

	return &Tracer{
		trace:          trace,
		store:          store,
		contextType:    reflect.TypeOf((*context.Context)(nil)).Elem(),
		traceTagType:   reflect.TypeOf((*EvalTraceTag)(nil)).Elem(),
		funcAttrsCache: map[string]int{},
		builtinsMap:    builtinFunctionsMap,
		eval: patcherEvaluationPhaseFields{
			closureEvalCount: map[TraceStoreKey]int{},
		},
		firstPhasePatcher: firstPhasePatcher{
			trace:                trace,
			mapper:               mapper,
			structFieldBaseNodes: structFieldBaseNodes,
		},
		secondPhasePatcher: secondPhasePatcher{
			patching:             patcherPatchingPhaseFields{},
			mapper:               mapper,
			structFieldBaseNodes: &structFieldBaseNodes,
		},
	}
}

func (t *Tracer) traceInteger(value any, info *IntegerNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*integerEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*integerEvalResult]{
			tag:         info.tag,
			evalResults: []*integerEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&integerEvalResult{
			value: value,
		})

	return value
}

func (t *Tracer) traceFloat(value any, info *FloatNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*floatEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*floatEvalResult]{
			tag:         info.tag,
			evalResults: []*floatEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&floatEvalResult{
			value: value,
		})

	return value
}

func (t *Tracer) traceCall(out any, info *CallNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*callEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*callEvalResult]{
			tag:         info.tag,
			evalResults: []*callEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	var ret = out

	entry.evalResults = append(
		entry.evalResults,
		&callEvalResult{
			output: ret,
		})

	return ret
}

func (t *Tracer) tracePredicate(ret any, info *PredicateNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*closureEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*closureEvalResult]{
			tag:         info.tag,
			evalResults: []*closureEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&closureEvalResult{
			output: ret,
		})

	strBuiltinTag := info.builtinTag.TraceStoreKey()

	if cnt, ok := t.eval.closureEvalCount[strBuiltinTag]; ok {
		t.eval.closureEvalCount[strBuiltinTag] = cnt + 1
	} else {
		t.eval.closureEvalCount[strBuiltinTag] = 1
	}

	return ret
}

func (t *Tracer) traceBuiltin(output any, info *BuiltinNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*builtinEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*builtinEvalResult]{
			tag:         info.tag,
			evalResults: []*builtinEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	strBuiltinTag := info.tag.TraceStoreKey()
	cnt := t.eval.closureEvalCount[strBuiltinTag]
	t.eval.closureEvalCount[strBuiltinTag] = 0

	entry.evalResults = append(
		entry.evalResults,
		&builtinEvalResult{
			output:           output,
			closureEvalCount: cnt,
		},
	)

	return output
}

func (t *Tracer) traceBinary(output any, info *BinaryNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*binaryEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*binaryEvalResult]{
			tag:         info.tag,
			evalResults: []*binaryEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&binaryEvalResult{
			output: output,
		},
	)

	return output
}

func (t *Tracer) traceConditional(output any, info *ConditionalNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*conditionalEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*conditionalEvalResult]{
			tag:         info.tag,
			evalResults: []*conditionalEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&conditionalEvalResult{
			output: output,
		},
	)

	return output
}

func (t *Tracer) traceIdentifier(value any, info *IdentifierNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*identifierEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*identifierEvalResult]{
			tag:         info.tag,
			evalResults: []*identifierEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&identifierEvalResult{
			value: value,
		},
	)

	return value
}

func (t *Tracer) tracePointer(value any, info *PointerNodeTraceInfo) any {
	closureEntry, closureOk := traceEntryByTag[*closureEvalResult](t.trace, info.closureTag)
	if !closureOk {
		closureEntry = &traceEntry[*closureEvalResult]{
			tag:         info.closureTag,
			evalResults: []*closureEvalResult{},
		}
		t.trace.AddTrace(info.closureTag, closureEntry)
	}

	entry, ok := traceEntryByTag[*pointerEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*pointerEvalResult]{
			tag:         info.tag,
			evalResults: []*pointerEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	closureCallCount := len(closureEntry.evalResults)

	if len(entry.evalResults) > 0 && entry.evalResults[len(entry.evalResults)-1].closureCallCount == closureCallCount {
		return value
	}

	entry.evalResults = append(
		entry.evalResults,
		&pointerEvalResult{
			value:            value,
			closureCallCount: closureCallCount,
		},
	)

	return value
}

func (t *Tracer) traceVariableDeclarator(value any, info *VariableDeclaratorNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*variableDeclaratorEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*variableDeclaratorEvalResult]{
			tag:         info.tag,
			evalResults: []*variableDeclaratorEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&variableDeclaratorEvalResult{
			value: value,
		},
	)

	return value
}

func (t *Tracer) traceMember(value any, info *MemberNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*memberEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*memberEvalResult]{
			tag:         info.tag,
			evalResults: []*memberEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&memberEvalResult{
			//property: property,
			value:    value,
			optional: info.memberNode.Optional,
			method:   info.memberNode.Method,
		},
	)

	return value
}

func (t *Tracer) traceArray(values any, info *ArrayNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*arrayEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*arrayEvalResult]{
			tag:         info.tag,
			evalResults: []*arrayEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&arrayEvalResult{
			values: values,
		},
	)

	return values
}

func (t *Tracer) traceSlice(values any, info *SliceNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*sliceEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*sliceEvalResult]{
			tag:         info.tag,
			evalResults: []*sliceEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&sliceEvalResult{
			values: values,
		},
	)

	return values
}
func (t *Tracer) traceMap(value map[string]any, info *MapNodeTraceInfo) map[string]any {
	entry, ok := traceEntryByTag[*mapEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*mapEvalResult]{
			tag:         info.tag,
			evalResults: []*mapEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&mapEvalResult{
			value: value,
		},
	)

	return value
}

func (t *Tracer) tracePairValue(value any, info *PairNodeTraceInfo) any {
	entry, ok := traceEntryByTag[*pairEvalResult](t.trace, info.tag)
	if !ok {
		entry = &traceEntry[*pairEvalResult]{
			tag:         info.tag,
			evalResults: []*pairEvalResult{},
		}
		t.trace.AddTrace(info.tag, entry)
	}

	entry.evalResults = append(
		entry.evalResults,
		&pairEvalResult{
			value: value,
		},
	)

	return value
}

func (t *Tracer) Patches() []expr.Option {
	return []expr.Option{
		expr.Patch(&t.firstPhasePatcher),
		expr.Patch(&t.secondPhasePatcher),
		expr.AllowUndefinedVariables(), // Add this option to allow undefined variables like Bar in $env?.[Bar]
	}
}

func (p *firstPhasePatcher) Visit(node *ast.Node) {
	p.mapper.Visit(node)

	tag := buildTag(p.mapper, *node)

	switch typedNode := (*node).(type) {
	case *ast.CallNode:
		p.trace.AddTrace(tag, &traceEntry[*callEvalResult]{tag: tag})
	case *ast.PredicateNode:
		p.trace.AddTrace(tag, &traceEntry[*closureEvalResult]{tag: tag})
	case *ast.BuiltinNode:
		p.trace.AddTrace(tag, &traceEntry[*builtinEvalResult]{tag: tag})
	case *ast.BinaryNode:
		p.trace.AddTrace(tag, &traceEntry[*binaryEvalResult]{tag: tag})
	case *ast.ConditionalNode:
		p.trace.AddTrace(tag, &traceEntry[*conditionalEvalResult]{tag: tag})
	case *ast.IdentifierNode:
		p.trace.AddTrace(tag, &traceEntry[*identifierEvalResult]{tag: tag})
	case *ast.PointerNode:
		p.trace.AddTrace(tag, &traceEntry[*pointerEvalResult]{tag: tag})
	case *ast.VariableDeclaratorNode:
		p.trace.AddTrace(tag, &traceEntry[*variableDeclaratorEvalResult]{tag: tag})
	case *ast.MemberNode:
		p.trace.AddTrace(tag, &traceEntry[*memberEvalResult]{tag: tag})
		// Mark the base node of struct field accesses so they won't be patched.
		// In expr v1.17.7+, patching the base node causes compiler panics because
		// the Nature's structData is not properly maintained after patching.
		if !typedNode.Method {
			if nodeType := typedNode.Node.Type(); nodeType != nil {
				for nodeType.Kind() == reflect.Ptr {
					nodeType = nodeType.Elem()
				}
				if nodeType.Kind() == reflect.Struct {
					// Mark the base node (and all its children) as not patchable
					markStructFieldBaseNodes(typedNode.Node, p.structFieldBaseNodes)
				}
			}
		}
	case *ast.ArrayNode:
		p.trace.AddTrace(tag, &traceEntry[*arrayEvalResult]{tag: tag})
	case *ast.SliceNode:
		p.trace.AddTrace(tag, &traceEntry[*sliceEvalResult]{tag: tag})
	case *ast.MapNode:
		p.trace.AddTrace(tag, &traceEntry[*mapEvalResult]{tag: tag})
	case *ast.PairNode:
		p.trace.AddTrace(tag, &traceEntry[*pairEvalResult]{tag: tag})
	case *ast.BoolNode:
		p.trace.AddTrace(tag, &traceEntry[*boolEvalResult]{tag: tag})
	case *ast.IntegerNode:
		p.trace.AddTrace(tag, &traceEntry[*integerEvalResult]{tag: tag})
	case *ast.FloatNode:
		p.trace.AddTrace(tag, &traceEntry[*floatEvalResult]{tag: tag})
	}
}

func (p *secondPhasePatcher) Visit(node *ast.Node) {
	tag := buildTag(p.mapper, *node)

	// Skip patching if this node is marked as a struct field base node.
	// In expr v1.17.7+, patching such nodes causes compiler panics because
	// the Nature's structData is not properly maintained after patching.
	nodePtr := reflect.ValueOf(*node).Pointer()
	if (*p.structFieldBaseNodes)[nodePtr] {
		return
	}

	switch typedNode := (*node).(type) {
	case *ast.BoolNode:
		// this node cannot be patched
	case *ast.IntegerNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &IntegerNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, integerNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncInteger, args)
	case *ast.FloatNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &FloatNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, floatNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncFloat, args)
	case *ast.CallNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &CallNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, callNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncCall, args)
	case *ast.PredicateNode:
		ptrTracer := &PredicateNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, predicateNode: typedNode}

		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: ptrTracer})
		p.patchNode(node, &identifierTracerFuncPredicate, args)

		p.patching.builtinPredicateNodes = append(p.patching.builtinPredicateNodes, ptrTracer)

		if len(p.patching.closurePointerNodes) > 0 {
			for _, pointer := range p.patching.closurePointerNodes {
				pointer.closureTag = tag
			}
			p.patching.closurePointerNodes = p.patching.closurePointerNodes[:0]
		}
	case *ast.BuiltinNode:
		args := make([]ast.Node, 0, len(typedNode.Arguments)+2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &BuiltinNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, builtinNode: typedNode}})

		p.patchNode(node, &identifierTracerFuncBuiltin, args)

		if len(p.patching.builtinPredicateNodes) > 0 {
			for _, pointer := range p.patching.builtinPredicateNodes {
				pointer.builtinTag = tag
			}
			p.patching.builtinPredicateNodes = p.patching.builtinPredicateNodes[:0]
		}
	case *ast.BinaryNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &BinaryNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, binaryNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncBinary, args)
	case *ast.ConditionalNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &ConditionalNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, conditionalNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncConditional, args)
	case *ast.IdentifierNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &IdentifierNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, identifierNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncIdentifier, args)
	case *ast.PointerNode:
		ptrTracer := &PointerNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, pointerNode: typedNode}
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: ptrTracer})
		p.patchNode(node, &identifierTracerFuncPointer, args)
		p.patching.closurePointerNodes = append(p.patching.closurePointerNodes, ptrTracer)
	case *ast.VariableDeclaratorNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &VariableDeclaratorNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, variableDeclaratorNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncVariableDecorator, args)
	case *ast.MemberNode:
		if !typedNode.Method {
			args := make([]ast.Node, 0, 2)
			args = append(args, typedNode)
			args = append(args, &ast.ConstantNode{Value: &MemberNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, memberNode: typedNode}})
			p.patchNode(node, &identifierTracerFuncMember, args)
		}
	case *ast.ArrayNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &ArrayNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, arrayNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncArray, args)
	case *ast.SliceNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &SliceNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, sliceNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncSlice, args)
	case *ast.MapNode:
		args := make([]ast.Node, 0, 2)
		args = append(args, typedNode)
		args = append(args, &ast.ConstantNode{Value: &MapNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, mapNode: typedNode}})
		p.patchNode(node, &identifierTracerFuncMap, args)
	case *ast.PairNode:
		info := &PairNodeTraceInfo{baseTraceInfo: baseTraceInfo{tag: tag}, pairNode: typedNode}

		{
			args := make([]ast.Node, 0, 2)
			args = append(args, typedNode)
			args = append(args, &ast.ConstantNode{Value: info})
			p.patchNode(node, &identifierTracerFuncPairValue, args)
		}
	case *ast.ChainNode:
		// Skip patching ChainNode to avoid errors with identifiers like Bar in $env?.[Bar]
		// This will make the behavior closer to the normal processing without patches
		// We still have trace information from firstPhasePatcher
	}
}

func (p *secondPhasePatcher) patchNode(node *ast.Node, tracerCallee *ast.IdentifierNode, args []ast.Node) {
	patchNode := &ast.CallNode{
		Callee:    tracerCallee,
		Arguments: args,
	}
	patchNode.SetType((*node).Type())
	ast.Patch(node, patchNode)
}

// markStructFieldBaseNodes marks a node and all its descendants as struct field base nodes.
// These nodes should not be patched because in expr v1.17.7+, patching them causes compiler
// panics due to the Nature's structData not being properly maintained after patching.
func markStructFieldBaseNodes(node ast.Node, nodes map[uintptr]bool) {
	ast.Walk(&node, &structFieldBaseMarker{nodes: nodes})
}

type structFieldBaseMarker struct {
	nodes map[uintptr]bool
}

func (m *structFieldBaseMarker) Visit(node *ast.Node) {
	nodePtr := reflect.ValueOf(*node).Pointer()
	m.nodes[nodePtr] = true
}
