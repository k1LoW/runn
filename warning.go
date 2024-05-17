package runn

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

var deprecationWarnings sync.Map

func printDeprecationWarnings() {
	if os.Getenv("RUNN_DISABLE_DEPRECATION_WARNING") != "" {
		return
	}
	first := true
	if os.Getenv("GTIHUB_ACTIONS") != "" {
		deprecationWarnings.Range(func(key, value any) bool {
			if first {
				fmt.Println()
				first = false
			}
			fmt.Printf("::warning title=runn deprecation warning::%s\n", value)
			deprecationWarnings.Delete(key)
			return true
		})
		return
	}

	warningf := color.New(color.FgYellow).FprintfFunc()
	stderr := colorable.NewColorableStderr()
	deprecationWarnings.Range(func(key, value any) bool {
		if first {
			fmt.Println()
			first = false
		}
		warningf(stderr, "Deprecation warning: %s\n", value)
		deprecationWarnings.Delete(key)
		return true
	})
}
