// Package strictsyncprimitives flags sync.Mutex / sync.RWMutex /
// sync.WaitGroup / sync.Once values inside workflow code. Real OS-level
// synchronization primitives block the worker thread; the workflow
// scheduler is single-threaded, so they are also unnecessary.
package strictsyncprimitives

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictsyncprimitives",
	Doc:      "Bans sync.Mutex/RWMutex/WaitGroup/Once in workflow code; the workflow scheduler is single-threaded.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var banned = map[string]struct{}{
	"Mutex": {}, "RWMutex": {}, "WaitGroup": {}, "Once": {}, "Cond": {},
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
			switch e := node.(type) {
			case *ast.SelectorExpr:
				if isSyncType(pass, e) {
					pass.Report(analysis.Diagnostic{
						Pos: e.Pos(), End: e.End(),
						Message: "sync." + e.Sel.Name + " in workflow code; the workflow scheduler is single-threaded, no synchronization needed",
					})
				}
			}
			return true
		})
		return true
	})
	return nil, nil
}

func isSyncType(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	if _, ok := banned[sel.Sel.Name]; !ok {
		return false
	}
	if pass.TypesInfo != nil {
		if obj := pass.TypesInfo.ObjectOf(sel.Sel); obj != nil {
			if pkg := obj.Pkg(); pkg != nil && pkg.Path() == "sync" {
				if _, ok := obj.Type().(*types.Named); ok {
					return true
				}
			}
		}
	}
	if id, ok := sel.X.(*ast.Ident); ok && id.Name == "sync" {
		return true
	}
	return false
}
