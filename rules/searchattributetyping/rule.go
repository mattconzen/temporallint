// Package searchattributetyping flags workflow.UpsertSearchAttributes
// calls whose value map is `map[string]interface{}{...}`. Untyped
// search-attribute writes are a common source of incorrect indexing on
// Temporal Cloud; the typed setters
// (temporal.NewSearchAttributeKeyString etc.) are safer.
//
// Heuristic: we flag any UpsertSearchAttributes call whose argument is
// a map[string]interface{} composite literal, regardless of contents.
package searchattributetyping

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "searchattributetyping",
	Doc:      "UpsertSearchAttributes with map[string]interface{} loses type info; prefer typed search attribute keys.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#not-properly-scoping-semantic-workflow-ids",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		if !temporalctx.MatchSelectorCall(call, "workflow", "UpsertSearchAttributes") {
			return
		}
		if len(call.Args) < 2 {
			return
		}
		cl, ok := call.Args[1].(*ast.CompositeLit)
		if !ok {
			return
		}
		mt, ok := cl.Type.(*ast.MapType)
		if !ok {
			return
		}
		// Key must be string and value must be interface{} for the heuristic to fire.
		keyIsString := false
		if id, ok := mt.Key.(*ast.Ident); ok && id.Name == "string" {
			keyIsString = true
		}
		valIsAny := false
		if it, ok := mt.Value.(*ast.InterfaceType); ok && (it.Methods == nil || len(it.Methods.List) == 0) {
			valIsAny = true
		}
		if id, ok := mt.Value.(*ast.Ident); ok && id.Name == "any" {
			valIsAny = true
		}
		if !keyIsString || !valIsAny {
			return
		}
		pass.Report(analysis.Diagnostic{
			Pos: cl.Pos(), End: cl.End(),
			Message: "UpsertSearchAttributes argument is map[string]interface{}; use typed SearchAttributeKey* constructors",
		})
	})
	return nil, nil
}
