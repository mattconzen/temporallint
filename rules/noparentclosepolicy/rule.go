// Package noparentclosepolicy flags workflow.ChildWorkflowOptions
// composite literals that omit ParentClosePolicy. Without an explicit
// policy the SDK defaults to PARENT_CLOSE_POLICY_TERMINATE — children
// are killed when the parent closes — which is rarely what the author
// intended for a long-running child. Setting it explicitly forces the
// author to think about the lifecycle.
package noparentclosepolicy

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "noparentclosepolicy",
	Doc:      "ChildWorkflowOptions should set ParentClosePolicy explicitly; the implicit default terminates orphans.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-using-parentclosepolicy",
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
		if temporalctx.HasField(cl, "ParentClosePolicy") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: cl.Pos(), End: cl.End(),
			Message: "ChildWorkflowOptions has no ParentClosePolicy; the implicit default terminates the child when the parent closes",
		})
	})
	return nil, nil
}
