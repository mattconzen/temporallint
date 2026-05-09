// Package multipleinputpayloads flags workflow functions that take more
// than one payload parameter alongside the workflow.Context. Workflows
// should accept a single input struct so the payload schema can evolve
// (add fields without breaking history) and so signal/query handlers
// stay aligned. Variadic / multi-arg signatures lock the caller into a
// positional contract that's hard to change.
package multipleinputpayloads

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "multipleinputpayloads",
	Doc:      "Workflow functions should take workflow.Context plus at most one input struct.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#using-multiple-inputresponse-payloads",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if !temporalctx.IsWorkflowFunc(pass, fn) {
			return
		}
		count := 0
		for _, field := range fn.Type.Params.List {
			if len(field.Names) == 0 {
				count++
				continue
			}
			count += len(field.Names)
		}
		if count <= 2 {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: fn.Type.Params.Pos(), End: fn.Type.Params.End(),
			Message: fmt.Sprintf("workflow %s takes %d parameters; use a single input struct so the schema can evolve", fn.Name.Name, count),
		})
	})
	return nil, nil
}
