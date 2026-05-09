// Package toomanyactivitytypes flags workflow functions that reference
// more than N distinct activity functions. Workflows with many activity
// types tend to be doing too much — split them.
//
// HEURISTIC, default-on. Threshold via -max-activity-types flag.
package toomanyactivitytypes

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var (
	enabled   = true
	threshold = 8
)

var Analyzer = &analysis.Analyzer{
	Name:     "toomanyactivitytypes",
	Doc:      "Workflow references too many distinct activity types (heuristic; default threshold 8).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#doing-too-many-things-in-one-workflow",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "toomanyactivitytypes", true, "enable toomanyactivitytypes (default true)")
	Analyzer.Flags.IntVar(&threshold, "max-activity-types", threshold, "max distinct activity types per workflow")
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
		seen := map[string]struct{}{}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !temporalctx.MatchSelectorCall(call, "workflow", "ExecuteActivity") {
				return true
			}
			if len(call.Args) < 2 {
				return true
			}
			id, ok := call.Args[1].(*ast.Ident)
			if !ok {
				return true
			}
			seen[id.Name] = struct{}{}
			return true
		})
		if len(seen) > threshold {
			pass.Report(analysis.Diagnostic{
				Pos: fn.Pos(), End: fn.Name.End(),
				Message: fmt.Sprintf("workflow %s references %d distinct activity types (threshold %d); consider splitting", fn.Name.Name, len(seen), threshold),
			})
		}
		return true
	})
	return nil, nil
}
