package temporalctx

import (
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// IsZeroValue reports whether expr is a syntactically obvious zero value:
// nil, "", 0, false, an empty composite literal, or a typed-nil cast like
// (*T)(nil). Useful for "did this return-statement return both a non-zero
// value AND a non-nil error?" style rules — false negatives (reporting
// non-zero when the expression is actually zero) hurt the rule less than
// false positives.
func IsZeroValue(pass *analysis.Pass, expr ast.Expr) bool {
	if expr == nil {
		return true
	}
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "nil" {
			return true
		}
		// Constant declarations that resolve to a zero value via type info.
		if pass != nil && pass.TypesInfo != nil {
			if tv, ok := pass.TypesInfo.Types[expr]; ok && tv.Value != nil {
				switch tv.Value.Kind() {
				case constant.Int:
					if v, ok := constant.Int64Val(tv.Value); ok && v == 0 {
						return true
					}
				case constant.String:
					if constant.StringVal(tv.Value) == "" {
						return true
					}
				case constant.Bool:
					return !constant.BoolVal(tv.Value)
				}
			}
		}
	case *ast.BasicLit:
		switch e.Kind.String() {
		case "INT":
			return e.Value == "0"
		case "FLOAT":
			return e.Value == "0" || e.Value == "0.0"
		case "STRING":
			return e.Value == `""`
		}
	case *ast.CompositeLit:
		// Empty literal counts as zero.
		return len(e.Elts) == 0
	case *ast.CallExpr:
		// (*T)(nil) and similar typed-nil casts.
		if len(e.Args) == 1 {
			if id, ok := e.Args[0].(*ast.Ident); ok && id.Name == "nil" {
				return true
			}
		}
	case *ast.UnaryExpr:
		if e.Op.String() == "&" {
			if cl, ok := e.X.(*ast.CompositeLit); ok && len(cl.Elts) == 0 {
				return true
			}
		}
	}
	// Type-info catch-all: explicit nil interface
	if pass != nil && pass.TypesInfo != nil {
		if t := pass.TypesInfo.TypeOf(expr); t != nil {
			if _, ok := t.Underlying().(*types.Interface); ok {
				if id, ok := expr.(*ast.Ident); ok && id.Name == "nil" {
					return true
				}
			}
		}
	}
	return false
}

// LoopHasCancelCheck reports whether body (the body of a for/range loop)
// either selects on `<-ctx.Done()` or calls a heartbeat-style API. Used
// by the activity-loop rules to recognise cancellation-aware loops.
func LoopHasCancelCheck(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		switch x := n.(type) {
		case *ast.SelectStmt:
			for _, c := range x.Body.List {
				cc, ok := c.(*ast.CommClause)
				if !ok {
					continue
				}
				if hasDoneRecv(cc.Comm) {
					found = true
					return false
				}
			}
		case *ast.CallExpr:
			if matchSelectorCall(x, "activity", "RecordHeartbeat") {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func hasDoneRecv(s ast.Stmt) bool {
	switch x := s.(type) {
	case *ast.ExprStmt:
		return isDoneRecv(x.X)
	case *ast.AssignStmt:
		for _, rhs := range x.Rhs {
			if isDoneRecv(rhs) {
				return true
			}
		}
	}
	return false
}

func isDoneRecv(expr ast.Expr) bool {
	u, ok := expr.(*ast.UnaryExpr)
	if !ok || u.Op.String() != "<-" {
		return false
	}
	call, ok := u.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	return matchSelectorCall(call, "ctx", "Done")
}

func matchSelectorCall(call *ast.CallExpr, recv, name string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if sel.Sel.Name != name {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return id.Name == recv
}
