package exprtrace

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/expr-lang/expr/ast"
	"github.com/xlab/treeprint"
)

type specialLabel struct {
	text string
}

func (s *specialLabel) String() string {
	return s.text
}

var (
	labelNotEvaluated  = &specialLabel{"[not evaluated]"}
	labelNotSpecified  = &specialLabel{"[not specified]"}
	labelAbstractArray = &specialLabel{"[...]"}
	labelThreeDots     = &specialLabel{"..."}
)

type TreePrinter interface {
	EvalEnv() EvalEnv
	PeekNodeEvalResultOutput(node ast.Node) any
}

type TreePrinterOption func(*treePrinterConfig) error

type treePrinterConfig struct {
	onCallNodeHook func(tp TreePrinter, tree treeprint.Tree, callNode *ast.CallNode, callOutput any)
}

func WithOnCallNodeHook(fn func(tp TreePrinter, tree treeprint.Tree, callNode *ast.CallNode, callOutput any)) TreePrinterOption {
	return func(c *treePrinterConfig) error {
		c.onCallNodeHook = fn
		return nil
	}
}

func PrintTree(trace *EvalTraceStore, env EvalEnv, node ast.Node, options ...TreePrinterOption) (treeprint.Tree, error) {
	tree := treeprint.New()
	mapper := &TagMapper{
		ptrMap:  map[uintptr]int{},
		counter: 0,
	}
	mapper.build(node)
	walker := treePrinter{
		trace:   trace,
		env:     env,
		counter: map[TraceStoreKey]int{},
		mapper:  mapper,
		config:  &treePrinterConfig{},
	}

	for _, opt := range options {
		if err := opt(walker.config); err != nil {
			return nil, err
		}
	}

	walker.walk(node, tree, -1, "")
	return tree, nil
}

type treePrinter struct {
	trace         *EvalTraceStore
	env           EvalEnv
	counter       map[TraceStoreKey]int
	mapper        *TagMapper
	stringBuilder strings.Builder
	config        *treePrinterConfig
}

func (tp *treePrinter) formatLabel(prefix string, label fmt.Stringer, output any) string {
	sb := tp.stringBuilder
	sb.Reset()

	if prefix != "" {
		sb.WriteString(prefix)
		sb.WriteString(" ")
	}
	sb.WriteString(label.String())
	sb.WriteString(" => ")

	switch o := output.(type) {
	case *specialLabel:
		sb.WriteString(o.text)
	case string:
		sb.WriteRune('"')
		sb.WriteString(strings.ReplaceAll(o, `"`, `\\"`))
		sb.WriteRune('"')
	case int:
		sb.WriteString(strconv.Itoa(o))
	case int64:
		sb.WriteString(strconv.FormatInt(o, 10))
	case bool:
		sb.WriteString(strconv.FormatBool(o))
	default:
		jsonBytes, err := json.Marshal(output)
		if err != nil {
			sb.WriteString("?")
		} else {
			sb.Write(jsonBytes)
		}
	}

	return sb.String()
}

func (tp *treePrinter) callCounterByTag(tag EvalTraceTag) int {
	if cnt, ok := tp.counter[tag.TraceStoreKey()]; ok {
		return cnt
	} else {
		return -1
	}
}

func (tp *treePrinter) isPrimitiveNode(node ast.Node) bool {
	switch node.(type) {
	case *ast.NilNode:
		return true
	case *ast.BoolNode:
		return true
	case *ast.IntegerNode:
		return true
	case *ast.FloatNode:
		return true
	case *ast.StringNode:
		return true
	default:
		return false
	}
}

func (tp *treePrinter) isAllPrimitiveNode(nodes ...ast.Node) bool {
	for i := range nodes {
		if !tp.isPrimitiveNode(nodes[i]) {
			return false
		}
	}

	return true
}

func (tp *treePrinter) addNode(tree treeprint.Tree, index int, prefix string, value string) treeprint.Tree {
	formattedValue := value
	if prefix != "" {
		formattedValue = fmt.Sprintf("%s %s", prefix, formattedValue)
	}

	if index >= 0 {
		formattedValue = fmt.Sprintf("(%d) %s", index, value)
	}

	return tree.AddNode(formattedValue)
}

func (tp *treePrinter) addBranch(tree treeprint.Tree, index int, value treeprint.Value) treeprint.Tree {
	formattedValue := value
	if index >= 0 {
		formattedValue = fmt.Sprintf("(%d) %s", index, value)
	}

	return tree.AddBranch(formattedValue)
}

func (tp *treePrinter) walk(node ast.Node, print treeprint.Tree, labelIndex int, labelPrefix string) {
	switch n := node.(type) {
	case *ast.NilNode:
		tp.addNode(print, labelIndex, labelPrefix, n.String())
	case *ast.BoolNode:
		tp.addNode(print, labelIndex, labelPrefix, n.String())
	case *ast.IntegerNode:
		tp.addNode(print, labelIndex, labelPrefix, n.String())
	case *ast.FloatNode:
		tp.addNode(print, labelIndex, labelPrefix, n.String())
	case *ast.StringNode:
		formatted, _ := json.Marshal(n.Value)
		tp.addNode(print, labelIndex, labelPrefix, string(formatted))
	case *ast.ConstantNode:
		tp.addNode(print, labelIndex, labelPrefix, n.String())
	case *ast.IdentifierNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*identifierEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].value)
			tp.addBranch(print, labelIndex, label)
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.UnaryNode:
		tp.addNode(print, labelIndex, labelPrefix, n.String())
	case *ast.BinaryNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*binaryEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].output)
			branch := tp.addBranch(print, labelIndex, label)
			tp.walk(n.Left, branch, -1, "")
			tp.walk(n.Right, branch, -1, "")
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.ChainNode:
		tp.walk(n.Node, print, -1, "")
	case *ast.MemberNode:
		if !n.Method {
			traceEntry, cnt := traceEntryAndEvalCountByNode[*memberEvalResult](tp, node)
			if cnt >= 0 {
				label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].value)
				tp.addBranch(print, -1, label)
			} else {
				label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
				tp.addBranch(print, labelIndex, label)
			}
		}
	case *ast.SliceNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*sliceEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].values)
			branch := tp.addBranch(print, labelIndex, label)
			tp.walk(n.Node, branch, labelIndex, "")
			if n.From != nil {
				tp.walk(n.From, branch, labelIndex, "(from)")
			} else {
				tp.addNode(branch, labelIndex, "(from)", labelNotSpecified.text)
			}
			if n.To != nil {
				tp.walk(n.To, branch, labelIndex, "(to)")
			} else {
				tp.addNode(branch, labelIndex, "(to)", labelNotSpecified.text)
			}
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.CallNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*callEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].output)
			index := cnt
			if len(traceEntry.evalResults) == 1 {
				index = -1
			}

			branch := tp.addBranch(print, index, label)
			if hook := tp.config.onCallNodeHook; hook != nil {
				hook(tp, branch, n, traceEntry.evalResults[cnt].output)
			}

			if !tp.isAllPrimitiveNode(n.Arguments...) {
				for j := range n.Arguments {
					tp.walk(n.Arguments[j], branch, -1, "")
				}
			}
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.BuiltinNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*builtinEvalResult](tp, node)

		if cnt >= 0 {
			evalResult := traceEntry.evalResults[cnt]
			label := tp.formatLabel(labelPrefix, node, evalResult.output)
			branch := tp.addBranch(print, labelIndex, label)
			for i := range n.Arguments {
				switch n.Arguments[i].(type) {
				case *ast.ClosureNode:
					tp.walkClosureNode(n.Arguments[i], branch, -1, "", evalResult.closureEvalCount)
				default:
					tp.walk(n.Arguments[i], branch, -1, "")
				}
			}
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.ClosureNode:
		panic("closure node should not be walked by this method, use walkClosureNode() instead")
	case *ast.PointerNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*pointerEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].value)
			tp.addNode(print, labelIndex, "", label)
		} else {
			tp.addNode(print, labelIndex, labelPrefix, labelNotEvaluated.text)
		}
	case *ast.VariableDeclaratorNode:
		_, cnt := traceEntryAndEvalCountByNode[*variableDeclaratorEvalResult](tp, node)
		if cnt >= 0 {
			tp.walk(n.Value, print, labelIndex, fmt.Sprintf("let %s =", n.Name))
			tp.walk(n.Expr, print, -1, "")
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.ConditionalNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*conditionalEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].output)
			index := cnt
			if len(traceEntry.evalResults) == 1 {
				index = -1
			}
			branch := tp.addBranch(print, index, label)

			var condValue bool
			{
				condOutput := tp.PeekNodeEvalResultOutput(n.Cond)
				if t, ok := condOutput.(bool); ok {
					condValue = t
				} else {
					panic("unexpected type")
				}
			}

			tp.walk(n.Cond, branch, -1, "")

			if condValue {
				tp.walk(n.Exp1, branch, -1, "(?)")
				tp.addNode(branch, -1, "(:)", labelNotEvaluated.text)
			} else {
				tp.addNode(branch, -1, "(?)", labelNotEvaluated.text)
				tp.walk(n.Exp2, branch, -1, "(:)")
			}
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.ArrayNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*arrayEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].values)
			branch := tp.addBranch(print, labelIndex, label)
			if !tp.isAllPrimitiveNode(n.Nodes...) {
				for j := range n.Nodes {
					tp.walk(n.Nodes[j], branch, j, "")
				}
			}
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.MapNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*mapEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, node, traceEntry.evalResults[cnt].value)
			branch := tp.addBranch(print, -1, label)
			for i := range n.Pairs {
				tp.walk(n.Pairs[i], branch, -1, "")
			}
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	case *ast.PairNode:
		traceEntry, cnt := traceEntryAndEvalCountByNode[*pairEvalResult](tp, node)
		if cnt >= 0 {
			label := tp.formatLabel(labelPrefix, n.Key, traceEntry.evalResults[cnt].value)
			branch := tp.addBranch(print, -1, label)

			if !tp.isPrimitiveNode(n.Value) {
				tp.walk(n.Value, branch, -1, "")
			}
		} else {
			label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
			tp.addBranch(print, labelIndex, label)
		}
	default:
		panic(fmt.Sprintf("undefined node type (%T)", node))
	}
}

func (tp *treePrinter) walkClosureNode(node ast.Node, print treeprint.Tree, labelIndex int, labelPrefix string, numClosureCalls int) {
	closureNode, ok := node.(*ast.ClosureNode)
	if !ok {
		panic("closure node is expected")
	}

	traceEntry, cnt := traceEntryAndEvalCountByNode[*closureEvalResult](tp, node)
	if cnt >= 0 {
		label := tp.formatLabel(labelPrefix, node, labelAbstractArray)
		branch := tp.addBranch(print, labelIndex, label)

		for i := range numClosureCalls {
			if i > 0 {
				traceEntry, cnt = traceEntryAndEvalCountByNode[*closureEvalResult](tp, node)
			}
			x := tp.formatLabel("", labelThreeDots, traceEntry.evalResults[cnt].output)
			closureBranch := tp.addBranch(branch, i, x)

			tp.walk(closureNode.Node, closureBranch, -1, "")
		}
	} else {
		label := tp.formatLabel(labelPrefix, node, labelNotEvaluated)
		tp.addBranch(print, -1, label)
	}
}

func (tp *treePrinter) EvalEnv() EvalEnv {
	return tp.env
}

func (tp *treePrinter) PeekNodeEvalResultOutput(node ast.Node) any {
	tag := buildTag(tp.mapper, node)
	cnt := tp.callCounterByTag(tag) + 1
	t, _ := tp.trace.TraceEntryByTag(tag)

	switch typedNode := node.(type) {
	case *ast.NilNode:
		return nil
	case *ast.BoolNode:
		return typedNode.Value
	case *ast.IntegerNode:
		return typedNode.Value
	case *ast.FloatNode:
		return typedNode.Value
	case *ast.ConstantNode:
		return typedNode.Value
	case *ast.StringNode:
		return typedNode.Value
	default:
		return t.EvalResultByCallCount(cnt).Output()
	}
}

func traceEntryAndEvalCountByNode[T NodeEvalResult](tp *treePrinter, node ast.Node) (*traceEntry[T], int) {
	tag := buildTag(tp.mapper, node)
	entry, ok := traceEntryByTag[T](tp.trace, tag)
	if !ok {
		panic("trace entry not found")
	}
	cnt := tp.callCounterByTag(tag) + 1
	if len(entry.evalResults)-1 >= cnt {
		tp.counter[tag.TraceStoreKey()] = cnt
		return entry, cnt
	} else {
		return entry, -1
	}
}
