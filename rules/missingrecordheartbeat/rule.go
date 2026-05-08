// Package missingrecordheartbeat flags activities containing a long
// loop (more than a small statement count, or with a Sleep call) whose
// body does NOT call activity.RecordHeartbeat. Long activities that
// don't heartbeat can't be detected as stalled; Temporal won't reschedule
// them on worker death until their HeartbeatTimeout fires.
//
// HEURISTIC, default-on. Disable per-rule with
// `temporallint -missingrecordheartbeat=false ./...`.
package missingrecordheartbeat

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "missingrecordheartbeat",
	Doc:      "Long-running activity loops should call activity.RecordHeartbeat (heuristic; default on).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-sending-heartbeats-from-activities",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "missingrecordheartbeat", true, "enable missingrecordheartbeat (default true)")
}

const longLoopThreshold = 5 // statements in body before we consider it "long"

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
			if !isLongLoop(body) {
				return true
			}
			if hasHeartbeat(body) {
				return true
			}
			pass.Report(analysis.Diagnostic{
				Pos: pos, End: end,
				Message: "long-running activity loop without activity.RecordHeartbeat; the activity cannot be detected as stalled",
			})
			return true
		})
		return true
	})
	return nil, nil
}

func isLongLoop(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
	if len(body.List) >= longLoopThreshold {
		return true
	}
	// Bodies that contain time.Sleep / workflow.Sleep are also "long".
	hasSleep := false
	ast.Inspect(body, func(n ast.Node) bool {
		if hasSleep {
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
		if sel.Sel.Name == "Sleep" {
			hasSleep = true
			return false
		}
		return true
	})
	return hasSleep
}

func hasHeartbeat(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
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
	return found
}
