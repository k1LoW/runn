package runn

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
)

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
		var targetEvaluator *evaluator
		for _, evaluator := range evaluators {
			if strings.HasPrefix(v, evaluator.scheme) {
				targetEvaluator = evaluator
			}
		}

		if targetEvaluator == nil {
			return value, nil
		}

		p := v[len(targetEvaluator.scheme):]
		if !hasExts(p, targetEvaluator.exts) && !hasTemplateSuffix(p, targetEvaluator.exts) {
			return value, fmt.Errorf("unsupported file extension: %s", p)
		}
		if operationRoot != "" {
			p = filepath.Join(operationRoot, p)
		}
		byteArray, err := readFile(p)
		if err != nil {
			return value, fmt.Errorf("read external files error: %w", err)
		}
		if store != nil && hasTemplateSuffix(p, targetEvaluator.exts) {
			tmpl, err := template.New(p).Parse(string(byteArray))
			if err != nil {
				return value, fmt.Errorf("template parse error: %w", err)
			}
			buf := new(bytes.Buffer)
			if err := tmpl.Execute(buf, store); err != nil {
				return value, fmt.Errorf("template excute error: %w", err)
			}
			byteArray = buf.Bytes()
		}
		var evaluatedObj any
		if err := targetEvaluator.unmarshal(byteArray, &evaluatedObj); err != nil {
			return value, fmt.Errorf("unmarshal error: %w", err)
		}

		return evaluatedObj, nil
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
