package runn

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

const multiple = "*"

type evaluator struct {
	scheme    string
	exts      []string
	unmarshal func(data []byte, v any) error
}

var (
	jsonEvaluator = &evaluator{scheme: "json://", exts: []string{"json"}, unmarshal: json.Unmarshal}
	yamlEvaluator = &evaluator{scheme: "yaml://", exts: []string{"yml", "yaml"}, unmarshal: yaml.Unmarshal}

	evaluators = []*evaluator{
		jsonEvaluator,
		yamlEvaluator,
	}
)

func evaluateSchema(value any, operationRoot string, store map[string]any) (any, error) {
	switch v := value.(type) {
	case string:
		var e *evaluator
		for _, evaluator := range evaluators {
			if strings.HasPrefix(v, evaluator.scheme) {
				e = evaluator
			}
		}
		if e == nil {
			return value, nil
		}

		p := v[len(e.scheme):]
		if strings.Contains(p, "://") {
			return value, fmt.Errorf("invalid path: %s", v)
		}
		if !hasExts(p, e.exts) && !hasTemplateSuffix(p, e.exts) {
			return value, fmt.Errorf("unsupported file extension: %s", p)
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(operationRoot, p)
		}

		if strings.Contains(p, multiple) {
			base, pattern := doublestar.SplitPattern(p)
			fsys := os.DirFS(base)
			matches, err := doublestar.Glob(fsys, pattern)
			if err != nil {
				return value, fmt.Errorf("glob error: %w", err)
			}
			sort.Slice(matches, func(i, j int) bool { return matches[i] < matches[j] })
			var outs []any
			for _, m := range matches {
				out, err := evalutateFile(filepath.Join(base, m), store, e)
				if err != nil {
					return value, fmt.Errorf("evaluate file error: %w", err)
				}
				outs = append(outs, out)
			}
			return outs, nil
		} else {
			out, err := evalutateFile(p, store, e)
			if err != nil {
				return value, fmt.Errorf("evaluate file error: %w", err)
			}
			return out, nil
		}
	}

	return value, nil
}

func hasExts(p string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(p, fmt.Sprintf(".%s", ext)) {
			return true
		}
	}
	return false
}

func hasTemplateSuffix(p string, exts []string) bool {
	for _, ext := range exts {
		if strings.HasSuffix(p, fmt.Sprintf(".%s.template", ext)) {
			return true
		}
	}
	return false
}

func evalutateFile(p string, store map[string]any, e *evaluator) (any, error) {
	b, err := readFile(p)
	if err != nil {
		return nil, fmt.Errorf("read external files error: %w", err)
	}
	if store != nil && hasTemplateSuffix(p, e.exts) {
		tmpl, err := template.New(p).Parse(string(b))
		if err != nil {
			return nil, fmt.Errorf("template parse error: %w", err)
		}
		buf := new(bytes.Buffer)
		if err := tmpl.Execute(buf, store); err != nil {
			return nil, fmt.Errorf("template excute error: %w", err)
		}
		b = buf.Bytes()
	}
	var out any
	if err := e.unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return out, nil
}
