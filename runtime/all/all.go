// Package all is the registry of every runtime.Check shipped by
// temporallint. Mirrors the structure of tools/temporallint/all (which
// is the registry for static analyzers) so that cmd/gen-rules can merge
// both into RULES.md.
package all

import (
	"sort"

	"github.com/mattconzen/temporallint/runtime"
	"github.com/mattconzen/temporallint/runtime/checks/historybytes"
	"github.com/mattconzen/temporallint/runtime/checks/historyevents"
	"github.com/mattconzen/temporallint/runtime/checks/individualpayload"
	"github.com/mattconzen/temporallint/runtime/checks/noworkflowtimeout"
)

// Checks returns the deduplicated, name-sorted list of every runtime
// check. Used by runtime.Main and by cmd/gen-rules.
func Checks() []runtime.Check {
	out := []runtime.Check{
		historybytes.New(),
		historyevents.New(),
		individualpayload.New(),
		noworkflowtimeout.New(),
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}
