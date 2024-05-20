package exprtrace

import (
	"fmt"

	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/parser"
	"github.com/xlab/treeprint"
)

type EvalEnv map[string]any

type EvalResult struct {
	Output             any
	Trace              *EvalTraceStore
	Source             string
	Env                EvalEnv
	TreePrinterOptions []TreePrinterOption
}

func (e *EvalResult) OutputAsBool() bool {
	switch vv := e.Output.(type) {
	case bool:
		return vv
	default:
		return false
	}
}

func (e *EvalResult) FormatTraceTree(opts ...TreePrinterOption) (string, error) {
	parsed, err := parser.Parse(e.Source)
	if err != nil {
		return "", err
	}

	var tree *treeprint.Node
	var modTree *treeprint.Node

	if t, err := PrintTree(e.Trace, e.Env, parsed.Node, append(e.TreePrinterOptions, opts...)...); err != nil {
		return "", err
	} else if tn, ok := t.(*treeprint.Node); !ok {
		return "", fmt.Errorf("*treeprint.Node type assertion failed")
	} else {
		tree = tn
	}

	if len(tree.Nodes) == 1 {
		modTree = tree.Nodes[0]
		modTree.Root = nil
	} else {
		if t, ok := treeprint.New().(*treeprint.Node); ok {
			modTree = t
		} else {
			return "", fmt.Errorf("*treeprint.Node type assertion failed")
		}
		modTree.Nodes = append(modTree.Nodes, tree.Nodes...)
		for _, node := range modTree.Nodes {
			node.Root = modTree
		}
	}

	modTree.SetValue(fmt.Sprintf("%s\n│", e.Source))

	return modTree.String(), nil
}

type EvalTraceTag struct {
	NodeIndex int
}

func buildTag(mapper *TagMapper, node ast.Node) EvalTraceTag {
	return EvalTraceTag{NodeIndex: mapper.indexByNode(node)}
}

type TraceStoreKey int

type EvalTraceStore struct {
	store map[TraceStoreKey]TraceEntry
}

func NewStore() *EvalTraceStore {
	return &EvalTraceStore{
		store: map[TraceStoreKey]TraceEntry{},
	}
}

func (ts *EvalTraceStore) AddTrace(tag EvalTraceTag, payload TraceEntry) {
	ts.store[tag.TraceStoreKey()] = payload
}

func (ts *EvalTraceStore) TraceEntryByTag(tag EvalTraceTag) (TraceEntry, bool) {
	if t, ok := ts.store[tag.TraceStoreKey()]; ok {
		return t, true
	} else {
		return nil, false
	}
}

func traceEntryByTag[T NodeEvalResult](trace *EvalTraceStore, tag EvalTraceTag) (*traceEntry[T], bool) {
	if t, ok := trace.store[tag.TraceStoreKey()]; ok {
		if ret, ok := t.(*traceEntry[T]); ok {
			return ret, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}
}

func (t *EvalTraceTag) TraceStoreKey() TraceStoreKey {
	return TraceStoreKey(t.NodeIndex)
}
