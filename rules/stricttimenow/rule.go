// Package stricttimenow bans time.Now() inside Temporal workflow code.
// time.Now is non-deterministic across replays; use workflow.Now(ctx).
package stricttimenow

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "stricttimenow",
	Doc:      "Bans time.Now() inside workflow code; use workflow.Now(ctx) for deterministic time.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#using-system-time-instead-of-workflow-time",
	Requires: temporalctx.Requires(),
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	temporalctx.RunCallBans(pass, []temporalctx.CallBan{{
		Pkg:     "time",
		Func:    "Now",
		Message: "time.Now() is non-deterministic in workflow code; use workflow.Now(ctx) instead",
		SuggestedFix: func(p *analysis.Pass, call *ast.CallExpr) analysis.SuggestedFix {
			ctxName := workflowContextParamName(p, call)
			if ctxName == "" {
				return analysis.SuggestedFix{}
			}
			return temporalctx.SimpleReplaceFix(
				"replace time.Now() with workflow.Now(ctx)",
				call.Pos(), call.End(),
				"workflow.Now("+ctxName+")",
			)
		},
	}})
	return nil, nil
}

// workflowContextParamName returns the name of the first parameter of the
// enclosing function declaration if and only if that parameter is a
// workflow.Context. Returns "" otherwise — which suppresses the
// suggested fix (we won't auto-fix without knowing the ctx variable name).
func workflowContextParamName(pass *analysis.Pass, target ast.Node) string {
	for _, f := range pass.Files {
		fn := temporalctx.EnclosingFuncDecl(f, target)
		if fn == nil {
			continue
		}
		if !temporalctx.IsWorkflowFunc(pass, fn) {
			continue
		}
		if fn.Type == nil || fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
			continue
		}
		field := fn.Type.Params.List[0]
		if len(field.Names) == 0 {
			continue
		}
		return field.Names[0].Name
	}
	return ""
}
