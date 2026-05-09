// Package signaldrainmissing flags workflows that handle signals via
// workflow.GetSignalChannel but never call Selector.HasPending() to
// drain leftover signals before returning. Without a drain check,
// signals that arrived between "main loop exits" and "workflow
// completes" are dropped — Temporal records the signal in history but
// the handler never runs, and the sender has no way to know.
//
// HEURISTIC, default-on. The check is intentionally conservative: the
// presence of any HasPending call anywhere in the workflow body is
// taken as evidence that the author thought about draining. Disable
// with `-signaldrainmissing=false`.
package signaldrainmissing

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "signaldrainmissing",
	Doc:      "Workflows that read signals should drain Selector.HasPending() before returning (heuristic).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-draining-signals-before-completing",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "signaldrainmissing", true, "enable signaldrainmissing (default true)")
}

func run(pass *analysis.Pass) (any, error) {
	if !enabled {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if !temporalctx.IsWorkflowFunc(pass, fn) || fn.Body == nil {
			return
		}
		if !callsGetSignalChannel(fn.Body) {
			return
		}
		if hasHasPending(fn.Body) {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: fn.Pos(), End: fn.Name.End(),
			Message: "workflow reads signals via workflow.GetSignalChannel but never calls Selector.HasPending; signals delivered between the last loop iteration and return are lost",
		})
	})
	return nil, nil
}

func callsGetSignalChannel(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if temporalctx.MatchSelectorCall(call, "workflow", "GetSignalChannel") {
			found = true
			return false
		}
		return true
	})
	return found
}

func hasHasPending(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "HasPending" {
			return true
		}
		found = true
		return false
	})
	return found
}
