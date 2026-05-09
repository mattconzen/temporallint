// Package pollingloopwithsleep flags workflow code that uses
// `for { ...; workflow.Sleep(ctx, ...); ... }` to poll for state. The
// idiomatic pattern is to let signals or timers wake the workflow
// instead of busy-polling.
//
// HEURISTIC, default-on. False positives expected for legitimate
// periodic-task workflows; disable per-rule with
// `temporallint -pollingloopwithsleep=false ./...`.
package pollingloopwithsleep

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "pollingloopwithsleep",
	Doc:      "Polling loops in workflows should use signals or timers, not workflow.Sleep (heuristic; default on).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#writing-polling-loops-in-workflow-code",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "pollingloopwithsleep", true, "enable pollingloopwithsleep (default true)")
}

func run(pass *analysis.Pass) (any, error) {
	if !enabled {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Nodes([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node, push bool) bool {
		if !push {
			return false
		}
		fn := n.(*ast.FuncDecl)
		if !temporalctx.IsWorkflowFunc(pass, fn) || fn.Body == nil {
			return false
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			var body *ast.BlockStmt
			var pos = node.Pos()
			switch x := node.(type) {
			case *ast.ForStmt:
				body = x.Body
			case *ast.RangeStmt:
				body = x.Body
			default:
				return true
			}
			if !bodyCallsWorkflowSleep(body) {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: pos, End: pos + 3,
				Message: "polling loop in workflow uses workflow.Sleep; consider signals or workflow.NewTimer-driven selectors",
			})
			return true
		})
		return true
	})
	return nil, nil
}

func bodyCallsWorkflowSleep(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found || n == nil {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Sleep" {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if !ok || id.Name != "workflow" {
			return true
		}
		found = true
		return false
	})
	return found
}
