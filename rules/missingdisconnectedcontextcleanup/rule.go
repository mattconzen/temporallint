// Package missingdisconnectedcontextcleanup flags `defer` blocks in
// workflow code that call workflow.ExecuteActivity using the parent
// (potentially cancelled) context rather than first calling
// workflow.NewDisconnectedContext. When the workflow is cancelled, the
// deferred activity inherits the cancellation and never runs.
//
// HEURISTIC, default-on. Disable per-rule with
// `temporallint -missingdisconnectedcontextcleanup=false ./...`.
package missingdisconnectedcontextcleanup

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "missingdisconnectedcontextcleanup",
	Doc:      "Deferred ExecuteActivity must use workflow.NewDisconnectedContext or it inherits cancellation (heuristic).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-using-disconnected-context-for-cleanup",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "missingdisconnectedcontextcleanup", true, "enable missingdisconnectedcontextcleanup (default true)")
}

func run(pass *analysis.Pass) (any, error) {
	if !enabled {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Nodes([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node, push bool) bool {
		if !push {
			return false
		}
		fn := n.(*ast.FuncDecl)
		if !temporalctx.IsWorkflowFunc(pass, fn) || fn.Body == nil {
			return false
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			ds, ok := node.(*ast.DeferStmt)
			if !ok {
				return true
			}
			body := deferBody(ds)
			if body == nil {
				return true
			}
			if !bodyHasExecuteActivity(body) {
				return true
			}
			if bodyCallsNewDisconnected(body) {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: ds.Pos(), End: ds.End(),
				Message: "deferred workflow.ExecuteActivity uses the parent context; cancellation will prevent the cleanup activity from running. Use workflow.NewDisconnectedContext.",
			})
			return true
		})
		return true
	})
	return nil, nil
}

func deferBody(ds *ast.DeferStmt) *ast.BlockStmt {
	switch fn := ds.Call.Fun.(type) {
	case *ast.FuncLit:
		return fn.Body
	case *ast.SelectorExpr, *ast.Ident:
		// `defer cleanup()` — we can't see the body; conservatively skip.
		return nil
	default:
		_ = fn
		return nil
	}
}

func bodyHasExecuteActivity(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found || n == nil {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if temporalctx.MatchSelectorCall(call, "workflow", "ExecuteActivity") ||
			temporalctx.MatchSelectorCall(call, "workflow", "ExecuteLocalActivity") {
			found = true
			return false
		}
		return true
	})
	return found
}

func bodyCallsNewDisconnected(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found || n == nil {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if temporalctx.MatchSelectorCall(call, "workflow", "NewDisconnectedContext") {
			found = true
			return false
		}
		return true
	})
	return found
}
