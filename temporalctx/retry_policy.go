package temporalctx

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"
	"time"

	"golang.org/x/tools/go/analysis"
)

// IsRetryPolicy reports whether cl is a temporal.RetryPolicy composite literal.
func IsRetryPolicy(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	return matchNamed(pass, cl, "go.temporal.io/sdk/temporal", "RetryPolicy") ||
		selectorMatches(cl.Type, "temporal", "RetryPolicy")
}

// IsStartWorkflowOptions reports whether cl is a client.StartWorkflowOptions
// composite literal.
func IsStartWorkflowOptions(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	return matchNamed(pass, cl, "go.temporal.io/sdk/client", "StartWorkflowOptions") ||
		selectorMatches(cl.Type, "client", "StartWorkflowOptions")
}

// IsChildWorkflowOptions reports whether cl is a workflow.ChildWorkflowOptions
// composite literal.
func IsChildWorkflowOptions(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	return matchNamed(pass, cl, "go.temporal.io/sdk/workflow", "ChildWorkflowOptions") ||
		selectorMatches(cl.Type, "workflow", "ChildWorkflowOptions")
}

func matchNamed(pass *analysis.Pass, cl *ast.CompositeLit, pkgPath, name string) bool {
	if cl == nil || pass == nil || pass.TypesInfo == nil {
		return false
	}
	t := pass.TypesInfo.TypeOf(cl)
	if t == nil {
		return false
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	if named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == pkgPath && named.Obj().Name() == name
}

// FieldByName returns the KeyValueExpr in cl whose key is `name`, or nil if
// no such field is set. Use HasField when you only need a boolean.
func FieldByName(cl *ast.CompositeLit, name string) *ast.KeyValueExpr {
	if cl == nil {
		return nil
	}
	for _, el := range cl.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		if key.Name == name {
			return kv
		}
	}
	return nil
}

// IntLiteralValue returns the integer constant value of expr if it can be
// determined statically. Used by rules that need to inspect numeric fields
// like MaximumAttempts.
func IntLiteralValue(pass *analysis.Pass, expr ast.Expr) (int64, bool) {
	if pass == nil || pass.TypesInfo == nil {
		return 0, false
	}
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Value == nil {
		return 0, false
	}
	v, ok := constant.Int64Val(tv.Value)
	return v, ok
}

// DurationLiteralValue is a best-effort evaluator for time.Duration
// expressions like `time.Minute`, `5 * time.Second`, `2 * time.Hour`.
// Returns ok=false for anything it cannot evaluate.
func DurationLiteralValue(pass *analysis.Pass, expr ast.Expr) (time.Duration, bool) {
	if expr == nil {
		return 0, false
	}
	if pass != nil && pass.TypesInfo != nil {
		if tv, ok := pass.TypesInfo.Types[expr]; ok && tv.Value != nil {
			if v, ok := constant.Int64Val(tv.Value); ok {
				return time.Duration(v), true
			}
		}
	}
	// AST fallback: handle `time.X` selector, `N * time.X`, `time.X * N`.
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		if d, ok := timeUnit(e); ok {
			return d, true
		}
	case *ast.BinaryExpr:
		if e.Op.String() == "*" {
			lhs, lok := DurationLiteralValue(pass, e.X)
			rhs, rok := DurationLiteralValue(pass, e.Y)
			if lok && rok {
				return lhs * rhs, true
			}
			if lok {
				if n, ok := basicLitInt(e.Y); ok {
					return time.Duration(n) * lhs, true
				}
			}
			if rok {
				if n, ok := basicLitInt(e.X); ok {
					return time.Duration(n) * rhs, true
				}
			}
		}
	case *ast.BasicLit:
		if n, ok := basicLitInt(e); ok {
			return time.Duration(n), true
		}
	}
	return 0, false
}

func timeUnit(sel *ast.SelectorExpr) (time.Duration, bool) {
	id, ok := sel.X.(*ast.Ident)
	if !ok || id.Name != "time" {
		return 0, false
	}
	switch sel.Sel.Name {
	case "Nanosecond":
		return time.Nanosecond, true
	case "Microsecond":
		return time.Microsecond, true
	case "Millisecond":
		return time.Millisecond, true
	case "Second":
		return time.Second, true
	case "Minute":
		return time.Minute, true
	case "Hour":
		return time.Hour, true
	}
	return 0, false
}

func basicLitInt(expr ast.Expr) (int64, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind.String() != "INT" {
		return 0, false
	}
	if strings.HasPrefix(lit.Value, "-") {
		// not handled — these rules don't expect negative durations
		return 0, false
	}
	var n int64
	for _, r := range lit.Value {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int64(r-'0')
	}
	return n, true
}
