package runn

import (
	"fmt"
	"testing"
)

func noColor(t *testing.T) {
	t.Helper()
	green = fmt.Sprint
	cyan = fmt.Sprint
	yellow = fmt.Sprint
	red = fmt.Sprint
}
