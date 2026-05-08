// Package workflowretrypolicy flags client.StartWorkflowOptions that
// set RetryPolicy. Workflow-level retries are almost always wrong:
// retries should live inside activities (which have their own retry
// policy). Workflow retries cause the entire workflow to restart,
// which is rarely the intent.
package workflowretrypolicy

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "workflowretrypolicy",
	Doc:      "StartWorkflowOptions.RetryPolicy is almost always wrong — retry inside activities instead.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#using-workflow-retries",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsStartWorkflowOptions(pass, cl) {
			return
		}
		kv := temporalctx.FieldByName(cl, "RetryPolicy")
		if kv == nil {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: kv.Pos(), End: kv.End(),
			Message: "StartWorkflowOptions.RetryPolicy retries the entire workflow on failure; retry inside activities instead",
			SuggestedFixes: []analysis.SuggestedFix{{
				Message: "remove RetryPolicy from StartWorkflowOptions",
				TextEdits: []analysis.TextEdit{{
					Pos:     kv.Pos(),
					End:     kv.End(),
					NewText: nil,
				}},
			}},
		})
	})
	return nil, nil
}
