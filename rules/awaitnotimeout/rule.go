// Package awaitnotimeout flags workflow.Await calls without an
// adjacent workflow.NewTimer used to bound the wait. An unconditional
// Await with no timeout can stall the workflow indefinitely if the
// condition never becomes true.
package awaitnotimeout

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "awaitnotimeout",
	Doc:      "workflow.Await without a paired NewTimer can stall forever; prefer workflow.AwaitWithTimeout.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#deadlocking-when-workflow-canceled",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	if !packageUsesTimer(pass) {
		// Function-by-function; if package never declares a timer, every
		// Await is suspect.
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		if sel.Sel.Name != "Await" {
			return
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Name != "workflow" {
			return
		}
		// We only flag if the package never uses workflow.NewTimer or
		// workflow.AwaitWithTimeout — heuristic to keep noise down.
		if packageUsesTimer(pass) {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: call.Pos(), End: call.End(),
			Message: "workflow.Await without NewTimer / AwaitWithTimeout in the package can stall forever",
		})
	})
	return nil, nil
}

func packageUsesTimer(pass *analysis.Pass) bool {
	for _, f := range pass.Files {
		used := false
		ast.Inspect(f, func(n ast.Node) bool {
			if used || n == nil {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if temporalctx.MatchSelectorCall(call, "workflow", "NewTimer") ||
				temporalctx.MatchSelectorCall(call, "workflow", "AwaitWithTimeout") {
				used = true
				return false
			}
			return true
		})
		if used {
			return true
		}
	}
	return false
}
