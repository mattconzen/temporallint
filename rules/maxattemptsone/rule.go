// Package maxattemptsone flags temporal.RetryPolicy literals whose
// MaximumAttempts is set to exactly 1. That value silently disables
// retries, which is almost never what the author meant — they usually
// wanted a non-retryable error type instead.
package maxattemptsone

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "maxattemptsone",
	Doc:      "RetryPolicy.MaximumAttempts == 1 silently disables retries; remove the field or use NonRetryableErrorTypes.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#preventing-activity-retries",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsRetryPolicy(pass, cl) && !temporalctx.IsActivityOptions(pass, cl) {
			return
		}
		// For ActivityOptions, descend into RetryPolicy field if present.
		if temporalctx.IsActivityOptions(pass, cl) {
			kv := temporalctx.FieldByName(cl, "RetryPolicy")
			if kv == nil {
				return
			}
			// Allow either &temporal.RetryPolicy{...} or temporal.RetryPolicy{...}
			inner := unwrapPointer(kv.Value)
			if rp, ok := inner.(*ast.CompositeLit); ok {
				checkAttempts(pass, rp)
			}
			return
		}
		checkAttempts(pass, cl)
	})
	return nil, nil
}

func unwrapPointer(e ast.Expr) ast.Expr {
	if u, ok := e.(*ast.UnaryExpr); ok && u.Op.String() == "&" {
		return u.X
	}
	return e
}

func checkAttempts(pass *analysis.Pass, rp *ast.CompositeLit) {
	kv := temporalctx.FieldByName(rp, "MaximumAttempts")
	if kv == nil {
		return
	}
	v, ok := temporalctx.IntLiteralValue(pass, kv.Value)
	if !ok || v != 1 {
		return
	}
	pass.Report(analysis.Diagnostic{
		Pos: kv.Pos(), End: kv.End(),
		Message: "MaximumAttempts: 1 silently disables retry; remove this field (default = unlimited) or mark specific errors as NonRetryableErrorTypes",
		SuggestedFixes: []analysis.SuggestedFix{{
			Message: "remove MaximumAttempts: 1",
			TextEdits: []analysis.TextEdit{{
				Pos: kv.Pos(),
				End: kv.End(),
				NewText: nil,
			}},
		}},
	})
}
