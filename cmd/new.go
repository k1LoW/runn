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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/k1LoW/runn"
	"github.com/k1LoW/runn/capture"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "create new runbook",
	Long:  `create new runbook.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("interactive mode is planned, but not yet implemented")
		}
		ctx := context.Background()
		rb := runn.NewRunbook(desc)
		if err := rb.AppendStep(args...); err != nil {
			return err
		}
		var (
			o   *os.File
			err error
		)
		if out == "" {
			o = os.Stdout
		} else {
			o, err = os.Create(filepath.Clean(out))
			if err != nil {
				return err
			}
			defer func() {
				if err := o.Close(); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
					os.Exit(1)
				}
			}()
		}

		fn := func(o *os.File) error {
			enc := yaml.NewEncoder(o)
			if err := enc.Encode(rb); err != nil {
				return err
			}
			return nil
		}

		if andRun {
			if err := runAndCapture(ctx, o, fn); err != nil {
				return err
			}
		} else {
			if err := fn(o); err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&desc, "desc", "", "", "description of runbook")
	newCmd.Flags().StringVarP(&out, "out", "", "", "output path of runbook")
	newCmd.Flags().BoolVarP(&andRun, "and-run", "", false, "run created runbook and capture the response for test")
}

func runAndCapture(ctx context.Context, o *os.File, fn func(*os.File) error) error {
	const newf = "new.yml"
	td, err := os.MkdirTemp("", "runn")
	if err != nil {
		return err
	}
	defer os.RemoveAll(td)
	tf, err := os.Create(filepath.Clean(filepath.Join(td, newf)))
	if err != nil {
		return err
	}
	defer func() {
		if err := tf.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}()

	if err := fn(tf); err != nil {
		return err
	}

	oo, err := runn.New(runn.Book(tf.Name()), runn.Capture(capture.Runbook(td)))
	if err != nil {
		return err
	}
	if err := oo.Run(ctx); err != nil {
		return err
	}

	entries, err := os.ReadDir(td)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() != newf {
			b, err := os.ReadFile(filepath.Join(td, e.Name()))
			if err != nil {
				return err
			}
			if _, err := o.Write(b); err != nil {
				return err
			}
		}
	}

	return nil
}
