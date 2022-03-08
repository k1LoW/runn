/*
Copyright © 2022 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/k1LoW/runn"
	"github.com/spf13/cobra"
)

var debug bool

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [FILE ...]",
	Short: "run books",
	Long:  `run books.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		total := 0
		failed := 0
		for _, p := range args {
			f, err := os.Stat(p)
			if err != nil {
				return err
			}
			paths := []string{}
			if f.IsDir() {
				entries, err := os.ReadDir(p)
				if err != nil {
					return err
				}
				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					paths = append(paths, filepath.Join(p, e.Name()))
				}
			} else {
				paths = append(paths, p)
			}
			for _, p := range paths {
				b, err := runn.LoadBookFile(p)
				if err != nil {
					continue
				}
				desc := b.Desc
				if desc == "" {
					desc = p
				}
				total += 1
				o, err := runn.New(runn.Book(p), runn.Debug(debug))
				if err != nil {
					fmt.Printf("%s ... %v\n", desc, err)
					failed += 1
					continue
				}
				if err := o.Run(ctx); err != nil {
					fmt.Printf("%s ... %v\n", desc, err)
					failed += 1
				} else {
					fmt.Printf("%s ... ok\n", desc)
				}
			}
		}
		fmt.Println("")
		var ts, fs string
		if total == 1 {
			ts = fmt.Sprintf("%d scenario", total)
		} else {
			ts = fmt.Sprintf("%d scenarios", total)
		}
		if failed == 1 {
			fs = fmt.Sprintf("%d failure", failed)
		} else {
			fs = fmt.Sprintf("%d failures", failed)
		}
		_, _ = fmt.Fprintf(os.Stdout, "%s, %s\n", ts, fs)
		if failed > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(debug, "debug", "", false, "debug")
}
