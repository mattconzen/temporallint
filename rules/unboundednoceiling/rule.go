// Package unboundednoceiling flags RetryPolicy literals that are
// effectively unbounded (MaximumAttempts == 0 or unset) without a
// MaximumInterval ceiling. Such a policy can retry forever with
// exponentially-growing intervals — a bug in disguise.
package unboundednoceiling

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "unboundednoceiling",
	Doc:      "RetryPolicy with unbounded MaximumAttempts requires a MaximumInterval ceiling.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#preventing-activity-retries",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsRetryPolicy(pass, cl) {
			return
		}
		// Unbounded == MaximumAttempts unset OR explicitly 0.
		if attempts := temporalctx.FieldByName(cl, "MaximumAttempts"); attempts != nil {
			if v, ok := temporalctx.IntLiteralValue(pass, attempts.Value); ok && v != 0 {
				return
			}
		}
		if temporalctx.HasField(cl, "MaximumInterval") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: cl.Pos(), End: cl.End(),
			Message: "RetryPolicy is unbounded (no MaximumAttempts) and has no MaximumInterval ceiling — retries can grow forever",
		})
	})
	return nil, nil
}
