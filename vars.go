package runn

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goccy/go-json"
)

const varsSupportScheme string = "json://"

func evaluateSchema(value interface{}, operationRoot string, store map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if !strings.HasPrefix(v, varsSupportScheme) {
			return value, nil
		}
		// json://
		fn := v[len(varsSupportScheme):]
		if operationRoot != "" {
			fn = filepath.Join(operationRoot, fn)
		}
		byteArray, err := readFile(fn)
		if err != nil {
			return value, fmt.Errorf("read external files error: %w", err)
		}
		if store != nil && strings.HasSuffix(fn, ".json.template") {
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
		var jsonObj interface{}
		if err := json.Unmarshal(byteArray, &jsonObj); err != nil {
			return value, fmt.Errorf("unmarshal error: %w", err)
		}

		return jsonObj, nil
	}

	return value, nil
}
