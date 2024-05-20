package builtin

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
)

func Time(v any) time.Time {
	switch vv := v.(type) {
	case string:
		t, err := dateparse.ParseStrict(vv)
		if err != nil {
			return t
		}
		return t
	default:
		t, err := dateparse.ParseStrict(fmt.Sprintf("%v", vv))
		if err != nil {
			return t
		}
		return t
	}
}
