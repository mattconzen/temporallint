// Command temporallint runs the temporallint static analyzers against
// Go packages, OR — when invoked with a `runtime` subcommand — performs
// runtime checks against a live Temporal server.
//
//	temporallint ./...                  # static analysis multichecker
//	temporallint runtime [flags...]     # runtime verification
//
// Argv dispatch is intentional: multichecker.Main consumes positional
// args as Go package patterns, so a regular flag-based subcommand
// dispatcher would conflict with `./...`. We instead peek at os.Args[1]
// before handing off.
package main

import (
	"os"

	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/mattconzen/monorepo/tools/temporallint/all"
	runtimecmd "github.com/mattconzen/monorepo/tools/temporallint/runtime/cmd"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "runtime" {
		os.Exit(runtimecmd.Main(os.Args[2:], os.Stdout, os.Stderr))
	}
	multichecker.Main(all.Analyzers()...)
}
