// Package nogracefuldrain flags `worker.Run(nil)` calls in main
// packages — running a worker without an interrupt channel means the
// process can't drain pending activity tasks on SIGTERM. Pass
// `worker.InterruptCh()` (or any context-derived channel) instead.
package nogracefuldrain

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "nogracefuldrain",
	Doc:      "worker.Run(nil) prevents graceful drain on shutdown; pass worker.InterruptCh() instead.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-draining-activity-tasks-before-shutdown",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg == nil || pass.Pkg.Name() != "main" {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Run" {
			return
		}
		// Receiver must look like a worker.Worker — heuristic: an
		// identifier returned from `worker.New(...)`. We don't trace data
		// flow, so simply require at least one argument and check whether
		// it's a literal nil.
		if len(call.Args) != 1 {
			return
		}
		id, ok := call.Args[0].(*ast.Ident)
		if !ok || id.Name != "nil" {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: call.Pos(), End: call.End(),
			Message: "worker.Run(nil) — no interrupt channel; pass worker.InterruptCh() to drain on shutdown",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "use worker.InterruptCh()",
				TextEdits: []analysis.TextEdit{{
					Pos:     call.Args[0].Pos(),
					End:     call.Args[0].End(),
					NewText: []byte("worker.InterruptCh()"),
				}},
			}},
		})
	})
	return nil, nil
}
