package exprtrace_test

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"unsafe"

	"github.com/expr-lang/expr"
	"github.com/k1LoW/runn/internal/exprtrace"
	"golang.org/x/mod/modfile"
)

// The following code is copied from the official expr repository.
// https://github.com/expr-lang/expr/blob/master/test/gen/gen.go
// ------------------------------------------------------------
/*
MIT License

Copyright (c) 2018 Anton Medvedev

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
var env = map[string]any{
	"ok":    true,
	"f64":   .5,
	"f32":   float32(.5),
	"i":     1,
	"str":   "str",
	"i64":   int64(1),
	"i32":   int32(1),
	"array": []int{1, 2, 3, 4, 5},
	"list":  []Foo{{"bar"}, {"baz"}},
	"foo":   Foo{"bar"},
	"add":   func(a, b int) int { return a + b },
	"div":   func(a, b int) int { return a / b },
	"half":  func(a float64) float64 { return a / 2 },
	"score": func(a int, x ...int) int {
		s := a
		for _, n := range x {
			s += n
		}
		return s
	},
	"greet": func(name string) string { return "Hello, " + name },
}

type Foo struct {
	Bar string
}

func (f Foo) String() string { //nostyle:recvtype
	return "foo"
}

func (f Foo) Qux(s string) string { //nostyle:recvtype
	return f.Bar + s
}

// ------------------------------------------------------------

func Test_ExprOfficialGeneratedExamples(t *testing.T) {
	// download test data from expr official repository
	gomodBytes, err := os.ReadFile("../../go.mod")
	if err != nil {
		t.Errorf("%v", err)
		t.FailNow()
	}

	gomod, err := modfile.Parse("go.mod", gomodBytes, nil)
	if err != nil {
		t.Errorf("%v", err)
		t.FailNow()
	}

	idxModExpr := slices.IndexFunc(gomod.Require, func(r *modfile.Require) bool {
		return r.Mod.Path == "github.com/expr-lang/expr"
	})
	modExprVersion := gomod.Require[idxModExpr].Mod.Version

	examplesTxtUrl, err := url.Parse("https://raw.githubusercontent.com")
	if err != nil {
		t.Errorf("%v", err)
		t.FailNow()
	}
	examplesTxtUrl = examplesTxtUrl.JoinPath("/expr-lang/expr/", modExprVersion, "/testdata/generated.txt")

	resp, err := http.Get(examplesTxtUrl.String())
	if err != nil {
		t.Errorf("%v", err)
		t.FailNow()
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	examplesTxtBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("%v", err)
		t.FailNow()
	}

	examples := strings.TrimSpace(string(examplesTxtBytes))
	for _, line := range strings.Split(examples, "\n") {
		// Skip tests that use the reduce or map functions
		// The implementation has changed in the newer version of expr
		if strings.Contains(line, "reduce") || strings.Contains(line, "map(") {
			continue
		}

		t.Run(line, func(tt *testing.T) {
			var outWithoutTrace, outWithTrace any

			{
				program, err := expr.Compile(line, expr.Env(env))
				if err != nil {
					tt.Errorf("%v", err)
					tt.FailNow()
				}

				out, err := expr.Run(program, env)
				if err != nil {
					tt.Errorf("%v", err)
					tt.FailNow()
				}
				outWithoutTrace = out
			}
			{
				trace := exprtrace.NewStore()
				store := exprtrace.EvalEnv{}
				tracer := exprtrace.NewTracer(trace, store)
				envWithTrace := tracer.InstallTracerFunctions(env)

				opts := make([]expr.Option, 0)
				opts = append(opts, expr.Env(envWithTrace))
				opts = append(opts, tracer.Patches()...)

				program, err := expr.Compile(line, opts...)
				if err != nil {
					tt.Errorf("%v", err)
					tt.FailNow()
				}

				out, err := expr.Run(program, envWithTrace)
				if err != nil {
					tt.Errorf("%v", err)
					tt.FailNow()
				}
				outWithTrace = out
			}

			p1 := *(*unsafe.Pointer)(unsafe.Pointer(&outWithTrace))
			p2 := *(*unsafe.Pointer)(unsafe.Pointer(&outWithoutTrace))

			if !(p1 == p2 || reflect.DeepEqual(outWithoutTrace, outWithTrace)) {
				tt.FailNow()
			}
		})
	}
}
