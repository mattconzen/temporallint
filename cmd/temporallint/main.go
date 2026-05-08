// Command temporallint is the standalone CLI for the temporallint static
// analysis suite. It runs every analyzer registered in the all package
// against the supplied Go packages.
//
//	temporallint ./...
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/mattconzen/monorepo/tools/temporallint/all"
)

func main() {
	multichecker.Main(all.Analyzers()...)
}
