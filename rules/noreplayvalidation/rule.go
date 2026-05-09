// Package noreplayvalidation flags packages containing workflow
// definitions that have no replay test. Replay tests catch
// non-deterministic changes (reordered statements, removed activity
// calls, new branches that didn't exist when the history was recorded)
// before the change reaches production. A workflow without a replay
// test loses one of Temporal's load-bearing safety nets.
//
// Default-OFF. Many repos isolate replay tests in a separate package
// (e.g. `replay_test/...`) so this rule's per-package scan would
// false-positive there. Enable with
// `temporallint -noreplayvalidation=true ./...`.
package noreplayvalidation

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = false

var Analyzer = &analysis.Analyzer{
	Name:     "noreplayvalidation",
	Doc:      "Packages with workflow definitions should also have a replay test (default-off).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-validating-replay-safety-before-deployments",
	Requires: temporalctx.Requires(),
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "noreplayvalidation", false, "enable noreplayvalidation (default false)")
}

func run(pass *analysis.Pass) (any, error) {
	if !enabled {
		return nil, nil
	}
	var firstWFPos *ast.FuncDecl
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if temporalctx.IsWorkflowFunc(pass, fn) {
				firstWFPos = fn
				break
			}
		}
		if firstWFPos != nil {
			break
		}
	}
	if firstWFPos == nil {
		return nil, nil
	}
	if hasReplayCall(pass) {
		return nil, nil
	}
	pass.Report(analysis.Diagnostic{
		Pos: firstWFPos.Pos(), End: firstWFPos.Name.End(),
		Message: "package contains workflow definitions but no replay test (worker.NewWorkflowReplayer / ReplayWorkflowHistory); non-deterministic changes will only surface in production",
	})
	return nil, nil
}

func hasReplayCall(pass *analysis.Pass) bool {
	for _, f := range pass.Files {
		found := false
		ast.Inspect(f, func(n ast.Node) bool {
			if found {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "NewWorkflowReplayer", "ReplayWorkflowHistory", "ReplayWorkflowHistoryFromJSONFile", "ReplayPartialWorkflowHistory":
				found = true
				return false
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}
