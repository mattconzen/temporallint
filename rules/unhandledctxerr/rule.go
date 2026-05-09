// Package unhandledctxerr flags workflow functions whose body contains
// at least one `Future.Get(ctx, ...)` call but never references
// `ctx.Err()`, `temporal.IsCanceledError`, or `workflow.IsContinueAsNew`.
// The heuristic catches workflows that propagate activity errors blindly
// without distinguishing cancellation from real failures.
//
// HEURISTIC, default-on. Disable with -unhandledctxerr=false.
package unhandledctxerr

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "unhandledctxerr",
	Doc:      "Workflow that calls Future.Get should handle ctx.Err()/cancellation explicitly (heuristic).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#deadlocking-when-workflow-canceled",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "unhandledctxerr", true, "enable unhandledctxerr (default true)")
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
		var firstGet *ast.CallExpr
		handlesCancel := false
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if firstGet == nil && isFutureGet(call) {
				firstGet = call
			}
			if temporalctx.MatchSelectorCall(call, "ctx", "Err") {
				handlesCancel = true
			}
			if temporalctx.MatchSelectorCall(call, "temporal", "IsCanceledError") {
				handlesCancel = true
			}
			return true
		})
		if firstGet != nil && !handlesCancel {
			pass.Report(analysis.Diagnostic{
				Pos: firstGet.Pos(), End: firstGet.End(),
				Message: "workflow uses Future.Get but never checks ctx.Err() or IsCanceledError; cancellation will be conflated with activity failures",
			})
		}
		return true
	})
	return nil, nil
}

// isFutureGet detects `<expr>.Get(ctx, ...)` calls. We can't tell from
// the AST alone whether the receiver is a workflow.Future, so this is
// inherently noisy; gated default-on per the heuristics policy.
func isFutureGet(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Get" {
		return false
	}
	if len(call.Args) < 1 {
		return false
	}
	id, ok := call.Args[0].(*ast.Ident)
	return ok && id.Name == "ctx"
}
