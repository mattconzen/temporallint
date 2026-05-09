// Package strictnethttp bans net/http client calls in workflow code.
// Network I/O must happen inside activities so retries / heartbeats /
// determinism work correctly.
package strictnethttp

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/temporallint/temporalctx"
)

var Analyzer = &analysis.Analyzer{
	Name:     "strictnethttp",
	Doc:      "Bans net/http client calls in workflow code; do I/O in activities.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#performing-network-calls-in-workflow-code",
	Requires: temporalctx.Requires(),
	Run: func(pass *analysis.Pass) (any, error) {
		bans := []temporalctx.CallBan{}
		for _, fn := range []string{"Get", "Post", "PostForm", "Head", "NewRequest", "NewRequestWithContext"} {
			bans = append(bans, temporalctx.CallBan{Pkg: "net/http", Func: fn,
				Message: "net/http call in workflow code; move I/O into an activity"})
		}
		temporalctx.RunCallBans(pass, bans)
		return nil, nil
	},
}
