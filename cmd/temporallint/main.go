// Command temporallint runs the temporallint static analyzers against
// Go packages, OR — when invoked with a known subcommand — runs the
// corresponding tool:
//
//	temporallint ./...                       # static analysis multichecker
//	temporallint runtime [flags...]          # runtime verification (Tier A)
//	temporallint cleanup-versions [flags...] # workflow.GetVersion cleanup (hybrid)
//	temporallint hook install|uninstall|status [flags...]  # git-hook integration
//
// Argv dispatch is intentional: multichecker.Main consumes positional
// args as Go package patterns, so a regular flag-based subcommand
// dispatcher would conflict with `./...`. We instead peek at os.Args[1]
// before handing off.
package main

import (
	"os"

	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/mattconzen/temporallint/all"
	cleanupcmd "github.com/mattconzen/temporallint/cleanupversions/cmd"
	"github.com/mattconzen/temporallint/hookcmd"
	runtimecmd "github.com/mattconzen/temporallint/runtime/cmd"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "runtime":
			os.Exit(runtimecmd.Main(os.Args[2:], os.Stdout, os.Stderr))
		case "cleanup-versions":
			os.Exit(cleanupcmd.Main(os.Args[2:], os.Stdout, os.Stderr))
		case "hook":
			os.Exit(hookcmd.Main(os.Args[2:], os.Stdout, os.Stderr))
		}
	}
	multichecker.Main(all.Analyzers()...)
}
