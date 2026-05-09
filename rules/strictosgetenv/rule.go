// Package strictosgetenv bans os.Getenv / os.LookupEnv / os.Environ in
// Temporal workflow code. Environment is non-deterministic across replays
// and across workers; pass configuration via workflow input or a
// well-known activity instead.
package strictosgetenv

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictosgetenv",
	Doc:      "Bans os.Getenv / os.LookupEnv / os.Environ in workflow code; pass config via workflow input.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#reading-environment-variables-in-workflow-code",
	Requires: temporalctx.Requires(),
	Run: func(pass *analysis.Pass) (any, error) {
		temporalctx.RunCallBans(pass, []temporalctx.CallBan{
			{Pkg: "os", Func: "Getenv", Message: "os.Getenv in workflow code is non-deterministic; pass config via workflow input"},
			{Pkg: "os", Func: "LookupEnv", Message: "os.LookupEnv in workflow code is non-deterministic; pass config via workflow input"},
			{Pkg: "os", Func: "Environ", Message: "os.Environ in workflow code is non-deterministic; pass config via workflow input"},
		})
		return nil, nil
	},
}
