// Package terminatevscancel flags client.TerminateWorkflow calls
// outside of test files. Termination skips graceful cancellation and
// any cleanup the workflow does in deferred blocks; CancelWorkflow is
// almost always the better choice.
package terminatevscancel

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "terminatevscancel",
	Doc:      "client.TerminateWorkflow skips graceful cleanup; prefer CancelWorkflow.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#terminating-rather-than-canceling",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.WithStack([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return false
		}
		call := n.(*ast.CallExpr)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "TerminateWorkflow" {
			return true
		}
		// Skip when invoked from a _test.go file; tests legitimately use it.
		if pass.Fset != nil {
			if pos := pass.Fset.Position(call.Pos()); strings.HasSuffix(pos.Filename, "_test.go") {
				return true
			}
		}
		pass.Report(analysis.Diagnostic{
			Pos: call.Pos(), End: call.End(),
			Message: "TerminateWorkflow skips graceful cancellation cleanup; prefer CancelWorkflow",
		})
		return true
	})
	return nil, nil
}
