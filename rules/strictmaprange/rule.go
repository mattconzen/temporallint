// Package strictmaprange flags `for k, v := range m` over a map type
// inside workflow code. Map iteration order in Go is randomized, which
// breaks workflow determinism. Sort the keys first.
package strictmaprange

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictmaprange",
	Doc:      "Flags `for range` over a map in workflow code; map iteration order is randomized.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#using-system-time-instead-of-workflow-time",
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
			rs, ok := node.(*ast.RangeStmt)
			if !ok {
				return true
			}
			if !isMapRange(pass, rs.X) {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: rs.Pos(), End: rs.For + 3, // highlight the `for` keyword span
				Message: "for range over a map in workflow code; map iteration order is randomized — sort keys first",
			})
			return true
		})
		return true
	})
	return nil, nil
}

func isMapRange(pass *analysis.Pass, expr ast.Expr) bool {
	if pass.TypesInfo == nil {
		return false
	}
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return false
	}
	if _, ok := t.Underlying().(*types.Map); ok {
		return true
	}
	return false
}
