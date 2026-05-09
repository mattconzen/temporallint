// Package missingworkflowtimeout flags client.StartWorkflowOptions
// composite literals that omit both WorkflowExecutionTimeout and
// WorkflowRunTimeout. Without one, the server applies a 10-year
// default — usually not what the author intended.
package missingworkflowtimeout

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "missingworkflowtimeout",
	Doc:      "StartWorkflowOptions should set WorkflowExecutionTimeout (or WorkflowRunTimeout); default is 10 years.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-setting-a-workflow-timeout",
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
		if temporalctx.HasField(cl, "WorkflowExecutionTimeout") || temporalctx.HasField(cl, "WorkflowRunTimeout") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: cl.Pos(), End: cl.End(),
			Message: "StartWorkflowOptions has neither WorkflowExecutionTimeout nor WorkflowRunTimeout; the server default of 10 years will apply",
		})
	})
	return nil, nil
}
