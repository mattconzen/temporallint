// Package stricttimesleep bans time.Sleep in workflow code. Real sleeps
// block the worker thread; use workflow.Sleep(ctx, d) which is recorded
// to history and survives replay.
package stricttimesleep

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "stricttimesleep",
	Doc:      "Bans time.Sleep in workflow code; use workflow.Sleep(ctx, d) instead.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#using-system-time-instead-of-workflow-time",
	Requires: temporalctx.Requires(),
	Run: func(pass *analysis.Pass) (any, error) {
		temporalctx.RunCallBans(pass, []temporalctx.CallBan{{
			Pkg:     "time",
			Func:    "Sleep",
			Message: "time.Sleep blocks the worker thread; use workflow.Sleep(ctx, d) so the timer is durable",
		}})
		return nil, nil
	},
}
