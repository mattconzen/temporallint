// Package preventretriesbytimeout flags ActivityOptions where the
// ScheduleToCloseTimeout is shorter than the RetryPolicy's
// InitialInterval — Temporal will never get a chance to retry because
// the overall deadline expires before the first backoff completes.
package preventretriesbytimeout

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "preventretriesbytimeout",
	Doc:      "ScheduleToCloseTimeout shorter than RetryPolicy.InitialInterval prevents retries.",
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
		stcKV := temporalctx.FieldByName(cl, "ScheduleToCloseTimeout")
		rpKV := temporalctx.FieldByName(cl, "RetryPolicy")
		if stcKV == nil || rpKV == nil {
			return
		}
		stc, stcOK := temporalctx.DurationLiteralValue(pass, stcKV.Value)
		if !stcOK {
			return
		}
		rp, ok := unwrapPointer(rpKV.Value).(*ast.CompositeLit)
		if !ok {
			return
		}
		initKV := temporalctx.FieldByName(rp, "InitialInterval")
		if initKV == nil {
			return
		}
		init, initOK := temporalctx.DurationLiteralValue(pass, initKV.Value)
		if !initOK {
			return
		}
		if stc < init {
			pass.Report(analysis.Diagnostic{
				Pos: stcKV.Pos(), End: stcKV.End(),
				Message: fmt.Sprintf("ScheduleToCloseTimeout (%s) is shorter than RetryPolicy.InitialInterval (%s); the activity will never retry", stc, init),
			})
		}
	})
	return nil, nil
}

func unwrapPointer(e ast.Expr) ast.Expr {
	if u, ok := e.(*ast.UnaryExpr); ok && u.Op.String() == "&" {
		return u.X
	}
	return e
}
