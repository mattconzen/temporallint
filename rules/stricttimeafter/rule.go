// Package stricttimeafter bans time.After / time.Tick / time.NewTicker in
// workflow code. Use workflow.NewTimer or workflow.Sleep instead.
package stricttimeafter

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "stricttimeafter",
	Doc:      "Bans time.After / time.Tick / time.NewTicker in workflow code.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#using-system-time-instead-of-workflow-time",
	Requires: temporalctx.Requires(),
	Run: func(pass *analysis.Pass) (any, error) {
		bans := []temporalctx.CallBan{}
		for _, fn := range []string{"After", "Tick", "NewTicker", "AfterFunc"} {
			bans = append(bans, temporalctx.CallBan{Pkg: "time", Func: fn,
				Message: "time." + fn + " in workflow code is non-deterministic; use workflow.NewTimer / workflow.Sleep"})
		}
		temporalctx.RunCallBans(pass, bans)
		return nil, nil
	},
}
