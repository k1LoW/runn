package runn

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/goccy/go-json"
)

const varsSupportScheme string = "json://"

func evaluateSchema(value interface{}, store map[string]interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, varsSupportScheme) {
			fn := v[len(varsSupportScheme):]
			byteArray, err := os.ReadFile(fn)
			if err != nil {
				return value, fmt.Errorf("read external files error: %w", err)
			}
			// If store exists, treat json as a template and replace it with store.
			if store != nil {
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
	}

	return value, nil
}
