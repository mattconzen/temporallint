// Package strictosexit bans os.Exit, log.Fatal*, log.Panic* in workflow
// code. A workflow that exits the process aborts replay and breaks every
// other workflow on the same worker.
package strictosexit

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictosexit",
	Doc:      "Bans os.Exit / log.Fatal / log.Panic in workflow code.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
	Requires: temporalctx.Requires(),
	Run: func(pass *analysis.Pass) (any, error) {
		bans := []temporalctx.CallBan{
			{Pkg: "os", Func: "Exit", Message: "os.Exit in workflow code aborts the worker; return an error instead"},
		}
		for _, fn := range []string{"Fatal", "Fatalf", "Fatalln", "Panic", "Panicf", "Panicln"} {
			bans = append(bans, temporalctx.CallBan{Pkg: "log", Func: fn,
				Message: "log." + fn + " in workflow code terminates the worker; return an error instead"})
		}
		temporalctx.RunCallBans(pass, bans)
		return nil, nil
	},
}
