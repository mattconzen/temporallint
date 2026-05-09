// Package activityignoresctxdone flags activity functions whose body
// contains a for/range loop that neither selects on `<-ctx.Done()` nor
// calls `activity.RecordHeartbeat`. Such loops cannot be cancelled
// promptly when the workflow asks the activity to stop.
//
// HEURISTIC, ships default-on. Disable per-rule with
// `temporallint -activityignoresctxdone=false ./...`. False positives
// occur when the loop body itself blocks on a context-aware operation
// (e.g. an HTTP client call with the activity context).
package activityignoresctxdone

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "activityignoresctxdone",
	Doc:      "Activity loops should select on <-ctx.Done() or call activity.RecordHeartbeat (heuristic; default on, disable with -activityignoresctxdone=false).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-sending-heartbeats-from-activities",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "activityignoresctxdone", true, "enable activityignoresctxdone (default true)")
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
		if !temporalctx.IsActivityFunc(pass, fn) || fn.Body == nil {
			return false
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if node == nil {
				return false
			}
			var body *ast.BlockStmt
			var pos, end = node.Pos(), node.End()
			switch x := node.(type) {
			case *ast.ForStmt:
				body = x.Body
				pos, end = x.Pos(), x.For+3
			case *ast.RangeStmt:
				body = x.Body
				pos, end = x.Pos(), x.For+3
			default:
				return true
			}
			if temporalctx.LoopHasCancelCheck(body) {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: pos, End: end,
				Message: "activity loop has no <-ctx.Done() select or activity.RecordHeartbeat call; cannot be cancelled promptly",
			})
			return true
		})
		return true
	})
	return nil, nil
}
