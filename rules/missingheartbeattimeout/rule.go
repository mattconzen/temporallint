// Package missingheartbeattimeout flags ActivityOptions that don't set
// HeartbeatTimeout in any package whose activities call
// activity.RecordHeartbeat. Without HeartbeatTimeout, missed heartbeats
// don't fail the activity — defeating the point of heartbeating.
//
// Heuristic: if ANY function in the package calls
// activity.RecordHeartbeat, every ActivityOptions in that package
// requires HeartbeatTimeout. False positives are possible when
// activities are split across packages.
package missingheartbeattimeout

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "missingheartbeattimeout",
	Doc:      "ActivityOptions in packages with heartbeating activities should set HeartbeatTimeout (heuristic).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-sending-heartbeats-from-activities",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	if !packageHeartbeats(pass) {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsActivityOptions(pass, cl) {
			return
		}
		if temporalctx.HasField(cl, "HeartbeatTimeout") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: cl.Pos(), End: cl.End(),
			Message: "package contains activity.RecordHeartbeat calls but this ActivityOptions has no HeartbeatTimeout",
		})
	})
	return nil, nil
}

// packageHeartbeats returns true when any file in the package contains a
// call to activity.RecordHeartbeat.
func packageHeartbeats(pass *analysis.Pass) bool {
	found := false
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if found {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "RecordHeartbeat" {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok || id.Name != "activity" {
				return true
			}
			found = true
			return false
		})
	}
	return found
}
