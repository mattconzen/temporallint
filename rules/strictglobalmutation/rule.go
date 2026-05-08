// Package strictglobalmutation flags writes to package-level variables
// from inside workflow code. Replays start with the package's initial
// state, so mutations from previous executions are invisible — every
// "shared" piece of state must live in workflow input or persisted state.
package strictglobalmutation

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictglobalmutation",
	Doc:      "Flags writes to package-level variables from workflow code.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#modifying-shared-state-in-workflow-code",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Nodes([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node, push bool) bool {
		if !push {
			return false
		}
		fn := n.(*ast.FuncDecl)
		if !temporalctx.IsWorkflowFunc(pass, fn) {
			return false
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			as, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, lhs := range as.Lhs {
				if isPackageLevelVar(pass, lhs) {
					pass.Report(analysis.Diagnostic{
						Pos: lhs.Pos(), End: lhs.End(),
						Message: "writing to a package-level variable from workflow code; replays start with initial state, so this mutation is lost",
					})
				}
			}
			return true
		})
		return true
	})
	return nil, nil
}

func isPackageLevelVar(pass *analysis.Pass, expr ast.Expr) bool {
	if pass.TypesInfo == nil {
		return false
	}
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	obj := pass.TypesInfo.ObjectOf(id)
	if obj == nil {
		return false
	}
	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}
	if v.Parent() == nil {
		return false
	}
	// Package-level variables have Pkg().Scope() as their parent.
	return v.Pkg() != nil && v.Parent() == v.Pkg().Scope()
}
