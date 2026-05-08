// Package signalhandlerblocksonactivity flags
// workflow.RegisterSignalHandler callbacks that synchronously call
// workflow.ExecuteActivity(...).Get(...). Blocking the signal handler
// goroutine on an activity stalls all other signals and the main
// workflow loop. Dispatch the work via a channel or workflow.Go.
package signalhandlerblocksonactivity

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "signalhandlerblocksonactivity",
	Doc:      "Signal handlers should not synchronously block on ExecuteActivity.Get; dispatch via channels or workflow.Go.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#assuming-signalsupdates-receive-in-order",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !temporalctx.MatchSelectorCall(call, "workflow", "RegisterSignalHandler") {
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
			if !ok || sel.Sel.Name != "Get" {
				return true
			}
			recvCall, ok := sel.X.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !temporalctx.MatchSelectorCall(recvCall, "workflow", "ExecuteActivity") &&
				!temporalctx.MatchSelectorCall(recvCall, "workflow", "ExecuteChildWorkflow") {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: c.Pos(), End: c.End(),
				Message: "signal handler synchronously waits on workflow.ExecuteActivity().Get; dispatch via a channel or workflow.Go",
			})
			return true
		})
	})
	return nil, nil
}
