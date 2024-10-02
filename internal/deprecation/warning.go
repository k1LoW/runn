package deprecation

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var warnings sync.Map

// AddWarning adds a deprecation warning message for a key.
func AddWarning(key, message string) {
	warnings.Store(key, message)
}

// PrintWarnings prints all deprecation warnings.
func PrintWarnings() {
	if os.Getenv("RUNN_DISABLE_DEPRECATION_WARNING") != "" {
		return
	}
	first := true
	if os.Getenv("GTIHUB_ACTIONS") != "" {
		warnings.Range(func(key, value any) bool {
			if first {
				fmt.Println()
				first = false
			}
			fmt.Printf("::warning title=runn deprecation warning::%s\n", value)
			warnings.Delete(key)
			return true
		})
		return
	}

	warningf := color.New(color.FgYellow).FprintfFunc()
	stderr := colorable.NewColorableStderr()
	warnings.Range(func(key, value any) bool {
		if first {
			fmt.Println()
			first = false
		}
		warningf(stderr, "Deprecation warning: %s\n", value)
		warnings.Delete(key)
		return true
	})
}
