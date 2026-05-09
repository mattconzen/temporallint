// Package workernotaskqueue flags `worker.New(client, "", ...)` calls
// where the task queue argument is the empty string. Without a task
// queue, the worker doesn't poll anything and silently does nothing.
package workernotaskqueue

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "workernotaskqueue",
	Doc:      "worker.New requires a non-empty task queue argument.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#starting-workflows-on-wrong-task-queue",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !temporalctx.MatchSelectorCall(call, "worker", "New") {
			return
		}
		if len(call.Args) < 2 {
			return
		}
		arg := call.Args[1]
		bl, ok := arg.(*ast.BasicLit)
		if !ok || bl.Kind.String() != "STRING" {
			return
		}
		if bl.Value != `""` && bl.Value != "``" {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: arg.Pos(), End: arg.End(),
			Message: "worker.New called with an empty task queue; the worker won't poll any tasks",
		})
	})
	return nil, nil
}
