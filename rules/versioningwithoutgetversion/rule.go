// Package versioningwithoutgetversion flags workflow functions whose
// body has multiple ExecuteActivity calls to differently-named
// activities in mutually exclusive if/else branches without a
// workflow.GetVersion guard. The pattern often indicates an unsafe
// version transition.
//
// HEURISTIC, default-on. Disable with -versioningwithoutgetversion=false.
// False positives are common — many legitimate branching workflows look
// like this. The rule's primary value is as a "look here when migrating"
// signal during code review.
package versioningwithoutgetversion

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "versioningwithoutgetversion",
	Doc:      "Branching workflows that call different activities should usually be guarded by workflow.GetVersion (heuristic).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-using-workflow-versioningpatching",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "versioningwithoutgetversion", true, "enable versioningwithoutgetversion (default true)")
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
			ifs, ok := node.(*ast.IfStmt)
			if !ok || ifs.Else == nil {
				return true
			}
			thenAct, _ := branchActivities(ifs.Body)
			var elseAct map[string]struct{}
			if eb, ok := ifs.Else.(*ast.BlockStmt); ok {
				elseAct, _ = branchActivities(eb)
			}
			if len(thenAct) == 0 || len(elseAct) == 0 {
				return true
			}
			// Different activities in then vs else? heuristic trigger
			differs := false
			for k := range thenAct {
				if _, ok := elseAct[k]; !ok {
					differs = true
					break
				}
			}
			if !differs {
				return true
			}
			if guardedByGetVersion(ifs) {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: ifs.Pos(), End: ifs.If + 2,
				Message: "if/else branches call different activities without workflow.GetVersion guard; if this is a version transition use GetVersion",
			})
			return true
		})
		return true
	})
	return nil, nil
}

func branchActivities(body *ast.BlockStmt) (map[string]struct{}, bool) {
	out := map[string]struct{}{}
	if body == nil {
		return out, false
	}
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if !temporalctx.MatchSelectorCall(call, "workflow", "ExecuteActivity") {
			return true
		}
		if len(call.Args) < 2 {
			return true
		}
		if id, ok := call.Args[1].(*ast.Ident); ok {
			out[id.Name] = struct{}{}
		}
		return true
	})
	return out, true
}

func guardedByGetVersion(ifs *ast.IfStmt) bool {
	// Recognise `if v := workflow.GetVersion(...); v == ...` shape.
	if ifs.Init == nil {
		// Try the condition itself.
		return condUsesGetVersion(ifs.Cond)
	}
	as, ok := ifs.Init.(*ast.AssignStmt)
	if !ok {
		return false
	}
	for _, rhs := range as.Rhs {
		if call, ok := rhs.(*ast.CallExpr); ok {
			if temporalctx.MatchSelectorCall(call, "workflow", "GetVersion") {
				return true
			}
		}
	}
	return false
}

func condUsesGetVersion(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found || n == nil {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if temporalctx.MatchSelectorCall(call, "workflow", "GetVersion") {
			found = true
			return false
		}
		return true
	})
	return found
}
