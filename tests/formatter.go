package tests

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"strings"
	"sync"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/formatters"
)

type kartuscheFormatter struct {
	*godog.ProgressFmt
	out            io.Writer
	cw             *capturingWriter
	mu             *sync.Mutex
	stepsTemplates *strings.Builder
}

func init() {
	formatters.Format("kartusche", "kartusche formatter", kartuscheFormatterFactory)
}

func kartuscheFormatterFactory(name string, w io.Writer) formatters.Formatter {
	cw := newCapturingWriter(w)
	return &kartuscheFormatter{
		ProgressFmt:    godog.NewProgressFmt(name, cw),
		out:            w,
		cw:             cw,
		mu:             new(sync.Mutex),
		stepsTemplates: new(strings.Builder),
	}
}

// this horrible code is due to godog's inadequate formatter interface.
func (k *kartuscheFormatter) Summary() {

	k.cw.startCapturing()
	k.ProgressFmt.Summary()

	captured := k.cw.buffer.String()
	idx := strings.Index(captured, "You can implement step definitions for undefined steps with these snippets:")
	if idx >= 0 {
		exprsText := captured[idx+len("You can implement step definitions for undefined steps with these snippets:"):]
		captured = captured[:idx]
		k.out.Write([]byte(captured))

		exprsText = strings.Replace(exprsText, "\x1b[0m\n\x1b[33m\n", "", 1)
		exprsText = strings.Replace(exprsText, "\n\x1b[0m", "", 1)

		af, err := parser.ParseFile(token.NewFileSet(), "foo.go", "package x\n"+exprsText, parser.SkipObjectResolution)
		if err != nil {
			panic(err)
		}

		numberOfArgs := map[string]int{}

		for _, d := range af.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if ok {

				if fd.Name.Name == "InitializeScenario" {
					for _, stmt := range fd.Body.List {
						ex, ok := stmt.(*ast.ExprStmt)
						if !ok {
							continue
						}
						ce, ok := ex.X.(*ast.CallExpr)
						if !ok {
							continue
						}
						bl, ok := ce.Args[0].(*ast.BasicLit)
						if !ok {
							continue
						}

						regexpString := strings.Trim(bl.Value, "`")

						fnId, ok := ce.Args[1].(*ast.Ident)
						cnt := numberOfArgs[fnId.Name]

						argList := []string{}
						for i := 0; i < cnt; i++ {
							argList = append(argList, fmt.Sprintf("arg%d", i+1))
						}

						fmt.Fprintf(k.out, `step('%s', (%s) => {
	throw new Error("not yet implemented")
})
`, regexpString, strings.Join(argList, ", "))

					}
				} else {
					cnt := 0
					for _, le := range fd.Type.Params.List {
						cnt += len(le.Names)
					}
					numberOfArgs[fd.Name.Name] = cnt
				}
			}
		}
	} else {
		k.out.Write([]byte(captured))
	}

}
