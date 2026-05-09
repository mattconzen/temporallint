// Package childworkflownotimeout flags workflow.ChildWorkflowOptions
// composite literals that omit both WorkflowExecutionTimeout and
// WorkflowRunTimeout — the same trap as parent-level
// StartWorkflowOptions, but for child workflows.
package childworkflownotimeout

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "childworkflownotimeout",
	Doc:      "ChildWorkflowOptions should set WorkflowExecutionTimeout or WorkflowRunTimeout.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-setting-a-workflow-timeout",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsChildWorkflowOptions(pass, cl) {
			return
		}
		if temporalctx.HasField(cl, "WorkflowExecutionTimeout") || temporalctx.HasField(cl, "WorkflowRunTimeout") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: cl.Pos(), End: cl.End(),
			Message: "ChildWorkflowOptions has neither WorkflowExecutionTimeout nor WorkflowRunTimeout",
		})
	})
	return nil, nil
}
