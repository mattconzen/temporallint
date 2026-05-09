// Package missingstarttoclosetimeout flags workflow.ActivityOptions
// composite literals that omit both StartToCloseTimeout and
// ScheduleToCloseTimeout. Without one of those, an activity that hangs
// forever will never be retried — Temporal has no other way to know it
// has stalled.
//
// This rule is the template / smoke rule for temporallint. Every other
// rule follows the same shape: rule.go (Analyzer), rule_test.go
// (analysistest.Run), testdata/src/violation, testdata/src/clean.
package missingstarttoclosetimeout

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "missingstarttoclosetimeout",
	Doc:      "ActivityOptions must set StartToCloseTimeout (or ScheduleToCloseTimeout); otherwise hung activities never retry.",
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
		if temporalctx.HasField(cl, "StartToCloseTimeout") || temporalctx.HasField(cl, "ScheduleToCloseTimeout") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos:     cl.Pos(),
			End:     cl.End(),
			Message: "ActivityOptions must set StartToCloseTimeout (or ScheduleToCloseTimeout) — without one, hung activities never retry",
		})
	})
	return nil, nil
}
