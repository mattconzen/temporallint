// Package strictmathrand bans math/rand and math/rand/v2 calls in
// workflow code. Random values must be generated via workflow.SideEffect
// (or via an activity) so they are recorded once and replayed
// deterministically.
package strictmathrand

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/monorepo/tools/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictmathrand",
	Doc:      "Bans math/rand calls in workflow code; use workflow.SideEffect to capture randomness deterministically.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#using-system-time-instead-of-workflow-time",
	Requires: temporalctx.Requires(),
	Run: func(pass *analysis.Pass) (any, error) {
		var bans []temporalctx.CallBan
		for _, fn := range []string{"Int", "Intn", "Int31", "Int31n", "Int63", "Int63n", "Float32", "Float64", "Read", "NewSource", "New", "Shuffle", "Perm"} {
			bans = append(bans,
				temporalctx.CallBan{Pkg: "math/rand", Func: fn,
					Message: "math/rand in workflow code is non-deterministic; wrap with workflow.SideEffect"},
				temporalctx.CallBan{Pkg: "math/rand/v2", Func: fn,
					Message: "math/rand/v2 in workflow code is non-deterministic; wrap with workflow.SideEffect"},
			)
		}
		temporalctx.RunCallBans(pass, bans)
		return nil, nil
	},
}
