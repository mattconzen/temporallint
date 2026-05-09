// Package startworkflowbadtaskqueue flags client.StartWorkflowOptions
// composite literals whose TaskQueue field is a string literal that
// doesn't match any task queue passed to worker.New in the same
// package. This catches the common typo where the workflow starter and
// the worker disagree on which queue to use.
//
// LIMITATION: scope is the analyzed package only. Cross-package wiring
// is the common production case, but multi-package analysis isn't
// available without a more sophisticated pipeline. We document this in
// the Doc string.
package startworkflowbadtaskqueue

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "startworkflowbadtaskqueue",
	Doc:      "StartWorkflowOptions.TaskQueue should match a worker.New(...) task queue in the same package.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#starting-workflows-on-wrong-task-queue",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	queues := workerNewTaskQueues(pass)
	if len(queues) == 0 {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsStartWorkflowOptions(pass, cl) {
			return
		}
		kv := temporalctx.FieldByName(cl, "TaskQueue")
		if kv == nil {
			return
		}
		bl, ok := kv.Value.(*ast.BasicLit)
		if !ok || bl.Kind.String() != "STRING" {
			return
		}
		if _, ok := queues[bl.Value]; ok {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: kv.Pos(), End: kv.End(),
			Message: fmt.Sprintf("TaskQueue %s does not match any worker.New(...) task queue in this package: %v", bl.Value, queueNames(queues)),
		})
	})
	return nil, nil
}

func workerNewTaskQueues(pass *analysis.Pass) map[string]struct{} {
	out := map[string]struct{}{}
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !temporalctx.MatchSelectorCall(call, "worker", "New") {
				return true
			}
			if len(call.Args) < 2 {
				return true
			}
			bl, ok := call.Args[1].(*ast.BasicLit)
			if !ok {
				return true
			}
			out[bl.Value] = struct{}{}
			return true
		})
	}
	return out
}

func queueNames(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
