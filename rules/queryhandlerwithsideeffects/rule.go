// Package queryhandlerwithsideeffects flags workflow.SetQueryHandler
// callbacks whose body calls workflow.ExecuteActivity, workflow.Sleep,
// workflow.NewTimer, or any other side-effecting workflow API. Query
// handlers must be pure functions over current state — Temporal calls
// them outside the deterministic execution flow.
package queryhandlerwithsideeffects

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "queryhandlerwithsideeffects",
	Doc:      "Query handlers must be pure; no ExecuteActivity / Sleep / NewTimer.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#querying-closed-workflows",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var bannedInQuery = map[string]struct{}{
	"ExecuteActivity":      {},
	"ExecuteLocalActivity": {},
	"ExecuteChildWorkflow": {},
	"Sleep":                {},
	"NewTimer":             {},
	"SideEffect":           {},
	"MutableSideEffect":    {},
	"ContinueAsNew":        {},
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !temporalctx.MatchSelectorCall(call, "workflow", "SetQueryHandler") {
			return
		}
		if len(call.Args) < 3 {
			return
		}
		fl, ok := call.Args[2].(*ast.FuncLit)
		if !ok {
			return
		}
		ast.Inspect(fl.Body, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			c, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := c.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok || id.Name != "workflow" {
				return true
			}
			if _, banned := bannedInQuery[sel.Sel.Name]; !banned {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: c.Pos(), End: c.End(),
				Message: "query handler calls workflow." + sel.Sel.Name + "; query handlers must be pure",
			})
			return true
		})
	})
	return nil, nil
}
