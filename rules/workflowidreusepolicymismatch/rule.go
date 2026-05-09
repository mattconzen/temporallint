// Package workflowidreusepolicymismatch flags client.StartWorkflowOptions
// composite literals that omit WorkflowIDReusePolicy. Default reuse
// policy on the server can vary; setting it explicitly avoids surprise
// behaviour when re-using a workflow ID for retries vs idempotency.
//
// HEURISTIC, default-on. Disable with -workflowidreusepolicymismatch=false.
package workflowidreusepolicymismatch

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var Analyzer = &analysis.Analyzer{
	Name:     "workflowidreusepolicymismatch",
	Doc:      "StartWorkflowOptions should set WorkflowIDReusePolicy explicitly (heuristic).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-properly-scoping-semantic-workflow-ids",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "workflowidreusepolicymismatch", true, "enable workflowidreusepolicymismatch (default true)")
}

func run(pass *analysis.Pass) (any, error) {
	if !enabled {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsStartWorkflowOptions(pass, cl) {
			return
		}
		if temporalctx.HasField(cl, "WorkflowIDReusePolicy") {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: cl.Pos(), End: cl.End(),
			Message: "StartWorkflowOptions has no WorkflowIDReusePolicy; set it explicitly to control reuse semantics",
		})
	})
	return nil, nil
}
