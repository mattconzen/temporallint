// Package unboundedloopnocnaw flags workflow functions whose bodies
// contain an unconditional `for { }` loop while the package never calls
// workflow.ContinueAsNew. Such workflows accumulate history events
// forever and eventually hit Temporal's hard size cap.
package unboundedloopnocnaw

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "unboundedloopnocnaw",
	Doc:      "Workflow with an unbounded for{} loop must call ContinueAsNew somewhere in the package.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-using-continueasnew",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	if packageCallsContinueAsNew(pass) {
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
			fs, ok := node.(*ast.ForStmt)
			if !ok {
				return true
			}
			if fs.Init != nil || fs.Cond != nil || fs.Post != nil {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: fs.Pos(), End: fs.For + 3,
				Message: "unbounded for{} loop in workflow without any workflow.ContinueAsNew in the package; history will grow forever",
			})
			return true
		})
		return true
	})
	return nil, nil
}

func packageCallsContinueAsNew(pass *analysis.Pass) bool {
	for _, f := range pass.Files {
		found := false
		ast.Inspect(f, func(n ast.Node) bool {
			if found || n == nil {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "ContinueAsNew" {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok || id.Name != "workflow" {
				return true
			}
			found = true
			return false
		})
		if found {
			return true
		}
	}
	return false
}
