// Package localactivitymisuse flags any use of workflow.LocalActivityOptions
// or workflow.ExecuteLocalActivity. Local activities have important
// limits — they don't appear in worker capacity tracking, can't heartbeat
// or be cancelled, run on the same worker as the workflow, and don't
// retry across worker restarts. Reach for regular activities first;
// only adopt local activities once you've measured the latency cost
// and accept the limits.
//
// Informational, default-on. Disable per-rule with
// `temporallint -localactivitymisuse=false ./...`.
package localactivitymisuse

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "localactivitymisuse",
	Doc:      "Flags workflow.LocalActivityOptions / workflow.ExecuteLocalActivity; reach for regular activities first.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#using-local-activities",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "localactivitymisuse", true, "enable localactivitymisuse (default true)")
}

func run(pass *analysis.Pass) (any, error) {
	if !enabled {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil), (*ast.CallExpr)(nil)}, func(n ast.Node) {
		switch x := n.(type) {
		case *ast.CompositeLit:
			if !temporalctx.IsLocalActivityOptions(pass, x) {
				return
			}
			pass.Report(analysis.Diagnostic{
				Pos: x.Pos(), End: x.End(),
				Message: "workflow.LocalActivityOptions in use; local activities skip retries across worker restarts and bypass capacity tracking — prefer regular activities",
			})
		case *ast.CallExpr:
			if !temporalctx.MatchSelectorCall(x, "workflow", "ExecuteLocalActivity") {
				return
			}
			pass.Report(analysis.Diagnostic{
				Pos: x.Pos(), End: x.End(),
				Message: "workflow.ExecuteLocalActivity in use; local activities skip retries across worker restarts and bypass capacity tracking — prefer regular activities",
			})
		}
	})
	return nil, nil
}
