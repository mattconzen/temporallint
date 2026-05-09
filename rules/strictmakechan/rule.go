// Package strictmakechan flags make(chan T) inside workflow code. Raw
// channels block the goroutine outside the deterministic scheduler; use
// workflow.NewChannel(ctx) instead.
package strictmakechan

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictmakechan",
	Doc:      "Bans make(chan T) in workflow code; use workflow.NewChannel(ctx) instead.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
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
		if !temporalctx.IsWorkflowFunc(pass, fn) {
			return false
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			id, ok := call.Fun.(*ast.Ident)
			if !ok || id.Name != "make" || len(call.Args) == 0 {
				return true
			}
			if _, ok := call.Args[0].(*ast.ChanType); !ok {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos:     call.Pos(),
				End:     call.End(),
				Message: "make(chan T) in workflow code blocks outside the deterministic scheduler; use workflow.NewChannel(ctx)",
			})
			return true
		})
		return true
	})
	return nil, nil
}
