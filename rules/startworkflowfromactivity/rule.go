// Package startworkflowfromactivity flags activity bodies that invoke
// client.ExecuteWorkflow or client.SignalWithStartWorkflow. Starting
// workflows from activities breaks the orchestration boundary; the
// workflow should start the child workflow itself.
package startworkflowfromactivity

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "startworkflowfromactivity",
	Doc:      "Activities should not start workflows; the parent workflow should orchestrate.",
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
			if node == nil {
				return false
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if sel.Sel.Name != "ExecuteWorkflow" && sel.Sel.Name != "SignalWithStartWorkflow" {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: call.Pos(), End: call.End(),
				Message: "activity should not call client." + sel.Sel.Name + "; let the parent workflow orchestrate",
			})
			return true
		})
		return true
	})
	return nil, nil
}
