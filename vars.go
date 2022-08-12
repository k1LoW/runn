package runn

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-json"
)

const varsSupportScheme string = "json://"

func evaluateSchema(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, varsSupportScheme) {
			byteArray, err := os.ReadFile(v[len(varsSupportScheme):])
			if err != nil {
				return value, fmt.Errorf("read external files error: %w", err)

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
