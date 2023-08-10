package runn

import (
	"fmt"
	"testing"
)

func noColor(t *testing.T) {
	t.Helper()
	currentGreen := green
	currentCyan := cyan
	currentYellow := yellow
	currentRed := red
	green = fmt.Sprint
	cyan = fmt.Sprint
	yellow = fmt.Sprint
	red = fmt.Sprint
	t.Cleanup(func() {
		green = currentGreen
		cyan = currentCyan
		yellow = currentYellow
		red = currentRed
	})
}
