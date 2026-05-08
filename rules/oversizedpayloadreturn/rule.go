// Package oversizedpayloadreturn flags activity return types whose
// declared struct has more than N fields. Large payloads bloat history
// size, slow down replay, and cost more on Temporal Cloud.
//
// HEURISTIC, default-on. Threshold configurable via -max-fields flag.
// False positives expected when a return struct legitimately holds
// many small fields (e.g. RPC response wrappers).
package oversizedpayloadreturn

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var (
	enabled   = true
	threshold = 16
)

var Analyzer = &analysis.Analyzer{
	Name:     "oversizedpayloadreturn",
	Doc:      "Activity return types with too many fields balloon workflow history (heuristic; default on).",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#passing-too-much-information-from-activities",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "oversizedpayloadreturn", true, "enable oversizedpayloadreturn (default true)")
	Analyzer.Flags.IntVar(&threshold, "max-fields", threshold, "max number of fields in an activity return struct")
}

func run(pass *analysis.Pass) (any, error) {
	if !enabled {
		return nil, nil
	}
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if !temporalctx.IsActivityFunc(pass, fn) {
			return
		}
		if fn.Type.Results == nil {
			return
		}
		for _, res := range fn.Type.Results.List {
			if count, name, ok := structFieldCount(pass, res.Type); ok && count > threshold {
				pass.Report(analysis.Diagnostic{
					Pos: res.Pos(), End: res.End(),
					Message: fmt.Sprintf("activity return type %s has %d fields (threshold %d); consider returning a smaller projection", name, count, threshold),
				})
			}
		}
	})
	return nil, nil
}

// structFieldCount returns the field count for the type referenced by
// expr if it resolves to a struct (named or anonymous, optionally
// pointer-wrapped), plus a friendly name for the diagnostic.
func structFieldCount(pass *analysis.Pass, expr ast.Expr) (int, string, bool) {
	if pass == nil || pass.TypesInfo == nil {
		return 0, "", false
	}
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return 0, "", false
	}
	for {
		ptr, ok := t.(*types.Pointer)
		if !ok {
			break
		}
		t = ptr.Elem()
	}
	name := "(anonymous struct)"
	if named, ok := t.(*types.Named); ok && named.Obj() != nil {
		name = named.Obj().Name()
	}
	if st, ok := t.Underlying().(*types.Struct); ok {
		return st.NumFields(), name, true
	}
	return 0, "", false
}
