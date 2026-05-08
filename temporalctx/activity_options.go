package temporalctx

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// IsActivityOptions reports whether cl is a composite literal of type
// go.temporal.io/sdk/workflow.ActivityOptions. It checks the resolved type
// first and falls back to a textual match for fixtures that don't have full
// type information available.
func IsActivityOptions(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	if cl == nil {
		return false
	}
	if pass != nil && pass.TypesInfo != nil {
		if t := pass.TypesInfo.TypeOf(cl); t != nil {
			if named, ok := t.(*types.Named); ok && named.Obj() != nil {
				obj := named.Obj()
				if obj.Pkg() != nil && obj.Pkg().Path() == "go.temporal.io/sdk/workflow" && obj.Name() == "ActivityOptions" {
					return true
				}
			}
		}
	}
	return selectorMatches(cl.Type, "workflow", "ActivityOptions")
}

// IsLocalActivityOptions is the local-activity variant.
func IsLocalActivityOptions(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	if cl == nil {
		return false
	}
	if pass != nil && pass.TypesInfo != nil {
		if t := pass.TypesInfo.TypeOf(cl); t != nil {
			if named, ok := t.(*types.Named); ok && named.Obj() != nil {
				obj := named.Obj()
				if obj.Pkg() != nil && obj.Pkg().Path() == "go.temporal.io/sdk/workflow" && obj.Name() == "LocalActivityOptions" {
					return true
				}
			}
		}
	}
	return selectorMatches(cl.Type, "workflow", "LocalActivityOptions")
}

// HasField reports whether cl has a key-value element with the given field
// name. It does NOT verify that the value is non-zero — a rule that cares
// about zero-vs-non-zero must check separately.
func HasField(cl *ast.CompositeLit, name string) bool {
	if cl == nil {
		return false
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
			return true
		}
	}
	return false
}
