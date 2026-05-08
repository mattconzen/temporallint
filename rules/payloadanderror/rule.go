// Package payloadanderror flags activity functions whose return
// statements provide BOTH a non-zero result AND a non-nil error. The
// caller can't tell whether to use the result or treat it as failed —
// either return a zero result with the error, or return the result with
// nil and signal failure via a sentinel/typed error value.
package payloadanderror

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "payloadanderror",
	Doc:      "Activity returns must not provide both a non-zero payload and a non-nil error.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#returning-both-payload-and-error",
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
		// Activity must return (T, error) — exactly two results.
		if fn.Type.Results == nil || len(fn.Type.Results.List) != 2 {
			return false
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			rs, ok := node.(*ast.ReturnStmt)
			if !ok || len(rs.Results) != 2 {
				return true
			}
			payload, errResult := rs.Results[0], rs.Results[1]
			if temporalctx.IsZeroValue(pass, payload) || temporalctx.IsZeroValue(pass, errResult) {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: rs.Pos(), End: rs.End(),
				Message: "activity returns both a non-zero payload and a non-nil error; the caller can't tell which to trust",
			})
			return true
		})
		return true
	})
	return nil, nil
}
