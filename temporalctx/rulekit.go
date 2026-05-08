package temporalctx

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// CallBan describes a "ban this qualified call inside workflow code" rule.
// Most non-determinism rules in Batch 2 are CallBan instances.
type CallBan struct {
	// Pkg is the import path (e.g. "time"), Func the function name
	// (e.g. "Now"). For methods on a value/type returned from a package
	// (rare here) leave Func empty and use a custom Match.
	Pkg, Func string
	// Message is the diagnostic text.
	Message string
	// SuggestedFix, when non-nil, is invoked with the offending CallExpr
	// and returns a single-edit fix. Returning a zero analysis.SuggestedFix
	// means "no fix is safe for this site".
	SuggestedFix func(pass *analysis.Pass, call *ast.CallExpr) analysis.SuggestedFix
}

// RunCallBans is the shared implementation: scan every workflow function
// (including reachable workflow-context-receiving helpers in the same
// package) for any call matching one of `bans`, then report.
func RunCallBans(pass *analysis.Pass, bans []CallBan) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Nodes([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node, push bool) bool {
		if !push {
			return false
		}
		fn := n.(*ast.FuncDecl)
		if !IsWorkflowFunc(pass, fn) {
			return false
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			for _, b := range bans {
				if matchQualifiedCall(pass, call, b.Pkg, b.Func) {
					reportCallBan(pass, call, b)
					return true
				}
			}
			return true
		})
		return true
	})
}

func reportCallBan(pass *analysis.Pass, call *ast.CallExpr, b CallBan) {
	d := analysis.Diagnostic{
		Pos:     call.Pos(),
		End:     call.End(),
		Message: b.Message,
	}
	if b.SuggestedFix != nil {
		fix := b.SuggestedFix(pass, call)
		if fix.Message != "" {
			d.SuggestedFixes = []analysis.SuggestedFix{fix}
		}
	}
	pass.Report(d)
}

// matchQualifiedCall returns true when call is `pkg.fn(...)` where pkg is
// the import path's identifier in the local file. It uses type info when
// available and falls back to import-name matching for testdata that
// stubs the temporal SDK packages.
func matchQualifiedCall(pass *analysis.Pass, call *ast.CallExpr, pkgPath, fnName string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != fnName {
		return false
	}
	// Type-info path
	if pass != nil && pass.TypesInfo != nil {
		if obj := pass.TypesInfo.ObjectOf(sel.Sel); obj != nil {
			if pkg := obj.Pkg(); pkg != nil && pkg.Path() == pkgPath {
				return true
			}
			// stdlib / package functions resolve as *types.Func with the
			// expected package; if not, fall through.
		}
		if id, ok := sel.X.(*ast.Ident); ok {
			if obj := pass.TypesInfo.ObjectOf(id); obj != nil {
				if pkgName, ok := obj.(*types.PkgName); ok && pkgName.Imported().Path() == pkgPath {
					return true
				}
			}
		}
	}
	// Fallback: match by short package name (last segment of import path)
	if id, ok := sel.X.(*ast.Ident); ok {
		short := pkgPath
		if i := strings.LastIndex(pkgPath, "/"); i >= 0 {
			short = pkgPath[i+1:]
		}
		if id.Name == short {
			return true
		}
	}
	return false
}

// SimpleReplaceFix replaces an entire node range with a fixed string.
func SimpleReplaceFix(message string, start, end token.Pos, replacement string) analysis.SuggestedFix {
	return analysis.SuggestedFix{
		Message: message,
		TextEdits: []analysis.TextEdit{{
			Pos:     start,
			End:     end,
			NewText: []byte(replacement),
		}},
	}
}

// Requires is the standard analyzer requirement: the inspector pass.
func Requires() []*analysis.Analyzer {
	return []*analysis.Analyzer{inspect.Analyzer}
}
