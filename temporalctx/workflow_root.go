// Package temporalctx provides shared helpers for analyzers that need to
// reason about Temporal-specific code shapes: which functions are workflows,
// which are activities, and where ActivityOptions composite literals appear.
package temporalctx

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// IsWorkflowFunc returns true when fn's first parameter has type
// workflow.Context (the canonical signal that a function is a Temporal
// workflow). It is intentionally conservative: helpers that take
// workflow.Context to access workflow APIs without being registered will
// also be flagged, which is the correct behaviour — anything reachable from
// a workflow root inherits workflow constraints.
func IsWorkflowFunc(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if fn == nil || fn.Type == nil || fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return false
	}
	first := fn.Type.Params.List[0]
	return isWorkflowContext(pass, first.Type)
}

// IsWorkflowFuncLit reports whether a function literal's first parameter is
// workflow.Context. Useful for detecting inline goroutine bodies invoked
// from workflows.
func IsWorkflowFuncLit(pass *analysis.Pass, fl *ast.FuncLit) bool {
	if fl == nil || fl.Type == nil || fl.Type.Params == nil || len(fl.Type.Params.List) == 0 {
		return false
	}
	first := fl.Type.Params.List[0]
	return isWorkflowContext(pass, first.Type)
}

func isWorkflowContext(pass *analysis.Pass, expr ast.Expr) bool {
	if pass != nil && pass.TypesInfo != nil {
		if t := pass.TypesInfo.TypeOf(expr); t != nil {
			named, ok := t.(*types.Named)
			if !ok {
				if ptr, ok := t.(*types.Pointer); ok {
					named, _ = ptr.Elem().(*types.Named)
				}
			}
			if named != nil && named.Obj() != nil {
				obj := named.Obj()
				if obj.Pkg() != nil && obj.Pkg().Path() == "go.temporal.io/sdk/workflow" && obj.Name() == "Context" {
					return true
				}
			}
		}
	}
	// Fallback for type-info-less analysis (e.g. generated fixtures): match
	// the source-level "workflow.Context" selector textually.
	return selectorMatches(expr, "workflow", "Context")
}

// IsActivityFunc returns true when fn's first parameter is context.Context
// AND the package imports go.temporal.io/sdk/activity OR the function is
// registered via worker.RegisterActivity in the same package.
//
// This is a heuristic: not every context.Context-first function is an
// activity. Callers should treat the result as advisory.
func IsActivityFunc(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if fn == nil || fn.Type == nil || fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return false
	}
	first := fn.Type.Params.List[0]
	if !isStdContext(pass, first.Type) {
		return false
	}
	return packageImportsTemporalActivity(pass) || isRegisteredAsActivity(pass, fn.Name.Name)
}

func isStdContext(pass *analysis.Pass, expr ast.Expr) bool {
	if pass != nil && pass.TypesInfo != nil {
		if t := pass.TypesInfo.TypeOf(expr); t != nil {
			if named, ok := t.(*types.Named); ok && named.Obj() != nil {
				obj := named.Obj()
				if obj.Pkg() != nil && obj.Pkg().Path() == "context" && obj.Name() == "Context" {
					return true
				}
			}
		}
	}
	return selectorMatches(expr, "context", "Context")
}

func packageImportsTemporalActivity(pass *analysis.Pass) bool {
	if pass == nil || pass.Pkg == nil {
		return false
	}
	for _, imp := range pass.Pkg.Imports() {
		if strings.HasPrefix(imp.Path(), "go.temporal.io/sdk/") {
			return true
		}
	}
	return false
}

func isRegisteredAsActivity(pass *analysis.Pass, fnName string) bool {
	if pass == nil {
		return false
	}
	found := false
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if found {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !strings.HasPrefix(sel.Sel.Name, "RegisterActivity") {
				return true
			}
			for _, arg := range call.Args {
				if id, ok := arg.(*ast.Ident); ok && id.Name == fnName {
					found = true
					return false
				}
			}
			return true
		})
	}
	return found
}

// EnclosingFuncDecl walks from a node back up to the FuncDecl that contains
// it, or returns nil if the node is at file scope. The pass.Inspector
// pre-order callback receives a parent stack — but most rules don't track
// that, so this helper does a fresh scan over the file.
func EnclosingFuncDecl(file *ast.File, target ast.Node) *ast.FuncDecl {
	if file == nil || target == nil {
		return nil
	}
	var found *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Pos() <= target.Pos() && target.End() <= fn.End() {
				found = fn
				return true
			}
		}
		return true
	})
	return found
}

func selectorMatches(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return id.Name == pkg && sel.Sel.Name == name
}
