// Package missingretrypolicy flags ActivityOptions composite literals
// that omit RetryPolicy. Without an explicit RetryPolicy, activities
// inherit the SDK default which may not be appropriate for the
// workload — long max-attempts, generous max-interval. Make the choice
// explicit.
package missingretrypolicy

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "missingretrypolicy",
	Doc:      "ActivityOptions should set RetryPolicy explicitly; default policy may not match workload.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#preventing-activity-retries",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsActivityOptions(pass, cl) {
			return
		}
		if temporalctx.HasField(cl, "RetryPolicy") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: cl.Pos(), End: cl.End(),
			Message: "ActivityOptions has no RetryPolicy; default may not match the workload — set RetryPolicy explicitly",
		})
	})
	return nil, nil
}
