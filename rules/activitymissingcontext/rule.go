// Package activitymissingcontext flags functions registered as
// activities (via worker.RegisterActivity) whose first parameter is not
// context.Context. The Temporal SDK requires this so it can plumb
// cancellation and metadata into activity execution.
package activitymissingcontext

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "activitymissingcontext",
	Doc:      "Functions registered as activities must accept context.Context as their first parameter.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-making-activities-idempotent",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	registered := registeredActivities(pass)
	if len(registered) == 0 {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Recv != nil {
			return
		}
		if _, ok := registered[fn.Name.Name]; !ok {
			return
		}
		if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
			pass.Report(analysis.Diagnostic{
				Pos: fn.Pos(), End: fn.Name.End(),
				Message: "activity " + fn.Name.Name + " has no parameters; it must accept context.Context first",
			})
			return
		}
		first := fn.Type.Params.List[0]
		if !isStdContext(pass, first.Type) {
			pass.Report(analysis.Diagnostic{
				Pos: first.Pos(), End: first.End(),
				Message: "activity " + fn.Name.Name + " must accept context.Context as its first parameter",
			})
		}
	})
	return nil, nil
}

func isStdContext(pass *analysis.Pass, expr ast.Expr) bool {
	if pass.TypesInfo != nil {
		if t := pass.TypesInfo.TypeOf(expr); t != nil {
			if named, ok := t.(*types.Named); ok && named.Obj() != nil {
				if pkg := named.Obj().Pkg(); pkg != nil && pkg.Path() == "context" && named.Obj().Name() == "Context" {
					return true
				}
			}
		}
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "context" && sel.Sel.Name == "Context"
}

// registeredActivities returns the names of identifiers passed to
// worker.RegisterActivity calls in the package.
func registeredActivities(pass *analysis.Pass) map[string]struct{} {
	out := map[string]struct{}{}
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !strings.HasPrefix(sel.Sel.Name, "RegisterActivity") {
				return true
			}
			for _, arg := range call.Args {
				if id, ok := arg.(*ast.Ident); ok {
					out[id.Name] = struct{}{}
				}
			}
			return true
		})
	}
	return out
}
