package builtin

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
)

func Time(v any) (time.Time, error) {
	switch vv := v.(type) {
	case string:
		return dateparse.ParseStrict(vv)
	default:
		return dateparse.ParseStrict(fmt.Sprintf("%v", vv))
	}
}
