package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/k1LoW/repin"
	"github.com/k1LoW/runn"
)

func main() {
	const repKey = "<!-- repin:fndoc -->"

	b, err := os.ReadFile("README.md")
	if err != nil {
		log.Fatal(err)
	}
	src := bytes.NewBuffer(b)
	rep := new(bytes.Buffer)
	out := new(bytes.Buffer)

	keys := []string{}
	for k := range runn.CDPFnMap {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		fn, ok := runn.CDPFnMap[k]
		if !ok {
			log.Fatalf("invalid key: %s", k)
		}
		as := ""
		if len(fn.Aliases) > 0 {
			as = fmt.Sprintf(" (aliases: `%s`)", strings.Join(fn.Aliases, "`, `"))
		}
		_, _ = fmt.Fprintf(rep, "**`%s`**%s\n\n", k, as)
		_, _ = fmt.Fprintf(rep, "%s\n\n", fn.Desc)
		_, _ = fmt.Fprint(rep, "```yaml\n")
		_, _ = fmt.Fprint(rep, "actions:\n")
		_, _ = fmt.Fprintf(rep, "  - %s:\n", k)
		for _, a := range fn.Args.ArgArgs() {
			_, _ = fmt.Fprintf(rep, "      %s: '%s'\n", a.Key, a.Example)
		}
		for _, a := range fn.Args.ResArgs() {
			_, _ = fmt.Fprintf(rep, "# record to current.%s:\n", a.Key)
		}
		_, _ = fmt.Fprint(rep, "```\n\n")

		if len(fn.Args.ArgArgs()) == 1 {
			_, _ = fmt.Fprint(rep, "or\n\n")
			_, _ = fmt.Fprint(rep, "```yaml\n")
			_, _ = fmt.Fprint(rep, "actions:\n")
			var e string
			for _, a := range fn.Args.ArgArgs() {
				e = a.Example
			}
			_, _ = fmt.Fprintf(rep, "  - %s: '%s'\n", k, e)
			_, _ = fmt.Fprint(rep, "```\n\n")
		}
	}

	if _, err := repin.Replace(src, rep, repKey, repKey, false, out); err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("README.md", out.Bytes(), os.ModePerm); err != nil {
		log.Fatal(err)
	}
}
