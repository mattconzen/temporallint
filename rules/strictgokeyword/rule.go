// Package strictgokeyword bans bare `go` statements inside workflow code.
// Native goroutines run outside the deterministic scheduler; use
// workflow.Go instead.
package strictgokeyword

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictgokeyword",
	Doc:      "Bans bare `go` statements in workflow code; use workflow.Go(ctx, fn) instead.",
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
			if gs, ok := node.(*ast.GoStmt); ok {
				pass.Report(analysis.Diagnostic{
					Pos:     gs.Pos(),
					End:     gs.End(),
					Message: "bare `go` statement in workflow code; use workflow.Go(ctx, fn) so the goroutine is part of the deterministic scheduler",
				})
			}
			return true
		})
		return true
	})
	return nil, nil
}
