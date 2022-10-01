package builtin

import (
	"encoding/json"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Compare(x, y interface{}, ignoreKeys ...string) bool {
	// normalize values
	bx, err := json.Marshal(x)
	if err != nil {
		return false
	}
	var vx interface{}
	if err := json.Unmarshal(bx, &vx); err != nil {
		return false
	}
	by, err := json.Marshal(y)
	if err != nil {
		return false
	}
	var vy interface{}
	if err := json.Unmarshal(by, &vy); err != nil {
		return false
	}

	diff := cmp.Diff(vx, vy, cmpopts.IgnoreMapEntries(func(key string, val interface{}) bool {
		for _, ignore := range ignoreKeys {
			if key == ignore {
				return true
			}
		}
		return false
	}))

	return diff == ""
}
