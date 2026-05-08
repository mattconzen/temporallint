// Package signalchanneloutsideselector flags workflow code that calls
// `workflow.GetSignalChannel(ctx, name).Receive(...)` directly (not
// inside a workflow.NewSelector). Direct Receive blocks the workflow
// until the signal arrives — usually wrong if the workflow has other
// work to do or other signals to handle. Use workflow.NewSelector to
// multiplex.
//
// HEURISTIC, default-on. Disable with -signalchanneloutsideselector=false.
package signalchanneloutsideselector

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "signalchanneloutsideselector",
	Doc:      "Direct GetSignalChannel(...).Receive blocks the workflow; use workflow.NewSelector for multiplexing (heuristic).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#assuming-signalsupdates-receive-in-order",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "signalchanneloutsideselector", true, "enable signalchanneloutsideselector (default true)")
}

func run(pass *analysis.Pass) (any, error) {
	if !enabled {
		return nil, nil
	}
	if packageUsesNewSelector(pass) {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		// Looking for `<something>.Receive(ctx, ...)`. Heuristic: if the
		// receiver is itself a workflow.GetSignalChannel(...) call, flag.
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Receive" {
			return
		}
		recvCall, ok := sel.X.(*ast.CallExpr)
		if !ok {
			return
		}
		if !temporalctx.MatchSelectorCall(recvCall, "workflow", "GetSignalChannel") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: call.Pos(), End: call.End(),
			Message: "GetSignalChannel(...).Receive used outside a workflow.NewSelector; signal handling will block the workflow",
		})
	})
	return nil, nil
}

func packageUsesNewSelector(pass *analysis.Pass) bool {
	for _, f := range pass.Files {
		used := false
		ast.Inspect(f, func(n ast.Node) bool {
			if used || n == nil {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if temporalctx.MatchSelectorCall(call, "workflow", "NewSelector") {
				used = true
				return false
			}
			return true
		})
		if used {
			return true
		}
	}
	return false
}
