// Package expensiveworkflowcomputation flags a small allowlist of
// expensive standard-library calls inside workflow code. Expensive work
// in a workflow has to re-run on every replay; even modest hashing or
// JSON encoding can tip a worker into running out of compute when many
// histories replay at once. Move these into activities so the result
// is checkpointed once.
//
// The ban list is intentionally narrow — this rule is not a general
// "is this CPU-heavy?" check, it's a list of well-known offenders:
//
//   - crypto/sha1, crypto/sha256, crypto/sha512, crypto/md5
//     (Sum, New)
//   - encoding/json (Marshal, Unmarshal)
//   - compress/gzip (NewReader, NewWriter)
//   - regexp (Compile, MustCompile)
//
// Default-on, gateable: `temporallint -expensiveworkflowcomputation=false ./...`.
package expensiveworkflowcomputation

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/temporallint/temporalctx"
)

var enabled = true

var bans = buildBans()

var Analyzer = &analysis.Analyzer{
	Name:     "expensiveworkflowcomputation",
	Doc:      "Bans a narrow allowlist of expensive stdlib calls (hashing, JSON, gzip, regexp.Compile) inside workflow code.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#performing-expensive-computation-in-workflow-code",
	Requires: temporalctx.Requires(),
	Run: func(pass *analysis.Pass) (any, error) {
		if !enabled {
			return nil, nil
		}
		temporalctx.RunCallBans(pass, bans)
		return nil, nil
	},
}

func init() {
	Analyzer.Flags.BoolVar(&enabled, "expensiveworkflowcomputation", true, "enable expensiveworkflowcomputation (default true)")
}

func buildBans() []temporalctx.CallBan {
	var out []temporalctx.CallBan
	hashFuncs := []string{"Sum", "New", "Sum224", "Sum256", "Sum384", "Sum512"}
	for _, hashPkg := range []string{"crypto/sha1", "crypto/sha256", "crypto/sha512", "crypto/md5"} {
		for _, fn := range hashFuncs {
			out = append(out, temporalctx.CallBan{Pkg: hashPkg, Func: fn,
				Message: hashPkg + "." + fn + " in workflow code; hashing replays on every history fetch — move into an activity"})
		}
	}
	for _, fn := range []string{"Marshal", "Unmarshal", "MarshalIndent"} {
		out = append(out, temporalctx.CallBan{Pkg: "encoding/json", Func: fn,
			Message: "encoding/json." + fn + " in workflow code; JSON encoding replays on every history fetch — move into an activity"})
	}
	for _, fn := range []string{"NewReader", "NewWriter"} {
		out = append(out, temporalctx.CallBan{Pkg: "compress/gzip", Func: fn,
			Message: "compress/gzip." + fn + " in workflow code; compression replays on every history fetch — move into an activity"})
	}
	for _, fn := range []string{"Compile", "MustCompile"} {
		out = append(out, temporalctx.CallBan{Pkg: "regexp", Func: fn,
			Message: "regexp." + fn + " in workflow code; pattern compilation replays on every history fetch — hoist to package scope or move into an activity"})
	}
	return out
}
