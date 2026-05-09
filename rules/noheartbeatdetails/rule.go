// Package noheartbeatdetails flags activity.RecordHeartbeat calls that
// pass no resumption details. A heartbeat without details signals
// liveness but doesn't checkpoint progress, so when the activity is
// retried after a worker crash it has to redo the work it had already
// completed. Pass at least one details argument so activity.GetInfo
// can resume.
package noheartbeatdetails

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "noheartbeatdetails",
	Doc:      "activity.RecordHeartbeat should pass resumption details so a retried activity can resume from where it left off.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-using-activity-heartbeat-details",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !temporalctx.MatchSelectorCall(call, "activity", "RecordHeartbeat") {
			return
		}
		// Args: ctx + zero or more details. <=1 means no details.
		if len(call.Args) > 1 {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: call.Pos(), End: call.End(),
			Message: "activity.RecordHeartbeat called without resumption details; a retried activity has no checkpoint to resume from",
		})
	})
	return nil, nil
}
