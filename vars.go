package runn

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/k1LoW/runn/internal/fs"
)

const multiple = "*"

type evaluator struct {
	scheme    string
	exts      []string
	unmarshal func(data []byte, v any) error
}

func textUnmarshal(data []byte, v any) error {
	s := string(data)
	if ptr, ok := v.(*any); ok {
		*ptr = s
		return nil
	}
	return fmt.Errorf("v must be a pointer to any")
}

var (
	jsonEvaluator = &evaluator{scheme: "json://", exts: []string{"json"}, unmarshal: json.Unmarshal}
	yamlEvaluator = &evaluator{scheme: "yaml://", exts: []string{"yml", "yaml"}, unmarshal: yaml.Unmarshal}
	fileEvaluator = &evaluator{scheme: "file://", exts: []string{"*"}, unmarshal: textUnmarshal}

	evaluators = []*evaluator{
		jsonEvaluator,
		yamlEvaluator,
		fileEvaluator,
	}
)

func evaluateSchema(value any, operationRoot string, store map[string]any) (any, error) {
	switch v := value.(type) {
	case string:
		var e *evaluator
		for _, evaluator := range evaluators {
			if strings.HasPrefix(v, evaluator.scheme) {
				e = evaluator
				break
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
			slices.Sort(matches)
			var outs []any
			for _, m := range matches {
				out, err := evaluateFile(filepath.Join(base, m), store, e)
				if err != nil {
					return value, fmt.Errorf("evaluate file error: %w", err)
				}
				outs = append(outs, out)
			}
			return outs, nil
		} else {
			out, err := evaluateFile(p, store, e)
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
		if ext == "*" {
			return true
		}
		if strings.HasSuffix(p, fmt.Sprintf(".%s", ext)) {
			return true
		}
	}
	return false
}

func hasTemplateSuffix(p string, exts []string) bool {
	for _, ext := range exts {
		if ext == "*" && strings.HasSuffix(p, ".template") {
			return true
		}
		if strings.HasSuffix(p, fmt.Sprintf(".%s.template", ext)) {
			return true
		}
	}
	return false
}

func evaluateFile(p string, store map[string]any, e *evaluator) (any, error) {
	b, err := fs.ReadFile(p)
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
