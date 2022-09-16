package runn

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Mainly respond to issue where the deep nested map becomes map[interface{}]interface{}
// and issue where yaml.MapSlice is defined as nested.

func yamlMarshal(in interface{}) ([]byte, error) {
	return yaml.Marshal(in)
}

func yamlUnmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

// normalize unmarshaled values
func normalize(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		res := make([]interface{}, len(v))
		for i, vv := range v {
			res[i] = normalize(vv)
		}
		return res
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, vv := range v {
			res[trimStringDelimiter(fmt.Sprintf("%v", k))] = normalize(vv)
		}
		return res
	case map[string]interface{}:
		res := make(map[string]interface{})
		for k, vv := range v {
			res[trimStringDelimiter(k)] = normalize(vv)
		}
		return res
	case []map[string]interface{}:
		res := make([]map[string]interface{}, len(v))
		for i, vv := range v {
			res[i] = toStrMap(vv)
		}
		return res
	case yaml.MapSlice:
		res := make(map[string]interface{})
		for _, i := range v {
			res[trimStringDelimiter(fmt.Sprintf("%v", i.Key))] = normalize(i.Value)
		}
		return res
	case string:
		return trimStringDelimiter(v)
	default:
		return v
	}
}

func toStrMap(v interface{}) map[string]interface{} {
	m := normalize(v)
	if m == nil {
		return map[string]interface{}{}
	}
	return m.(map[string]interface{})
}
