package builtin

import (
	"encoding/json"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Diff(x, y interface{}, ignoreKeys ...string) string {
	d, err := diff(x, y, ignoreKeys...)
	if err != nil {
		panic(err)
	}

	return d
}

func diff(x, y interface{}, ignoreKeys ...string) (string, error) {
	// normalize values
	bx, err := json.Marshal(x)
	if err != nil {
		return "", err
	}
	var vx interface{}
	if err := json.Unmarshal(bx, &vx); err != nil {
		return "", err
	}
	by, err := json.Marshal(y)
	if err != nil {
		return "", err
	}
	var vy interface{}
	if err := json.Unmarshal(by, &vy); err != nil {
		return "", err
	}

	diff := cmp.Diff(vx, vy, cmpopts.IgnoreMapEntries(func(key string, val interface{}) bool {
		for _, ignore := range ignoreKeys {
			if key == ignore {
				return true
			}
		}
		return false
	}))

	return diff, nil
}
