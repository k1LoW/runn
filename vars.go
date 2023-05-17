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
	prefix    string
	unmarshal func(data []byte, v interface{}) error
}

func (e evaluator) Scheme() string {
	return e.prefix + "://"
}

var (
	jsonEvaluator = &evaluator{prefix: "json", unmarshal: json.Unmarshal}
	yamlEvaluator = &evaluator{prefix: "yaml", unmarshal: yaml.Unmarshal}

	evaluators = []*evaluator{
		jsonEvaluator,
		yamlEvaluator,
	}
)

func evaluateSchema(value interface{}, operationRoot string, store map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		var targetEvaluator *evaluator
		for _, evaluator := range evaluators {
			if strings.HasPrefix(v, evaluator.Scheme()) {
				targetEvaluator = evaluator
			}
		}

		if targetEvaluator == nil {
			return value, nil
		}

		fn := v[len(targetEvaluator.Scheme()):]
		if operationRoot != "" {
			fn = filepath.Join(operationRoot, fn)
		}
		byteArray, err := readFile(fn)
		if err != nil {
			return value, fmt.Errorf("read external files error: %w", err)
		}
		if store != nil && strings.HasSuffix(fn, fmt.Sprintf(".%s.template", targetEvaluator.prefix)) {
			tmpl, err := template.New(fn).Parse(string(byteArray))
			if err != nil {
				return value, fmt.Errorf("template parse error: %w", err)
			}
			buf := new(bytes.Buffer)
			if err := tmpl.Execute(buf, store); err != nil {
				return value, fmt.Errorf("template excute error: %w", err)
			}
			byteArray = buf.Bytes()
		}
		var evaluatedObj interface{}
		if err := targetEvaluator.unmarshal(byteArray, &evaluatedObj); err != nil {
			return value, fmt.Errorf("unmarshal error: %w", err)
		}

		return evaluatedObj, nil
	}

	return value, nil
}
