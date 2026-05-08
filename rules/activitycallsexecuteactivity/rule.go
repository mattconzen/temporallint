// Package activitycallsexecuteactivity flags calls to
// workflow.ExecuteActivity / workflow.ExecuteLocalActivity from inside
// an activity function. Activities run on the worker, not in the
// workflow scheduler — they cannot orchestrate other activities the way
// workflows can.
package activitycallsexecuteactivity

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "activitycallsexecuteactivity",
	Doc:      "Activities cannot call workflow.ExecuteActivity / ExecuteLocalActivity; orchestration belongs in workflows.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#starting-workflows-from-activities",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Nodes([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node, push bool) bool {
		if !push {
			return false
		}
		fn := n.(*ast.FuncDecl)
		if !temporalctx.IsActivityFunc(pass, fn) || fn.Body == nil {
			return false
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if sel.Sel.Name != "ExecuteActivity" && sel.Sel.Name != "ExecuteLocalActivity" {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok || id.Name != "workflow" {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: call.Pos(), End: call.End(),
				Message: "workflow." + sel.Sel.Name + " called from inside an activity; activities cannot orchestrate other activities",
			})
			return true
		})
		return true
	})
	return nil, nil
}
