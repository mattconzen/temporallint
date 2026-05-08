// Package sideeffectnoresult flags workflow.SideEffect calls whose
// return value is discarded. SideEffect's whole purpose is to record a
// computed value into history so replay sees the same value — if you
// throw it away, you also throw away the determinism guarantee.
package sideeffectnoresult

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "sideeffectnoresult",
	Doc:      "workflow.SideEffect / MutableSideEffect return values must be captured; discarding them defeats the purpose.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-using-return-value-in-side-effects",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.ExprStmt)(nil)}, func(n ast.Node) {
		es := n.(*ast.ExprStmt)
		call, ok := es.X.(*ast.CallExpr)
		if !ok {
			return
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		if sel.Sel.Name != "SideEffect" && sel.Sel.Name != "MutableSideEffect" {
			return
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Name != "workflow" {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: call.Pos(), End: call.End(),
			Message: "workflow." + sel.Sel.Name + " result discarded; assign it and call .Get(...) to capture the recorded value",
		})
	})
	return nil, nil
}
