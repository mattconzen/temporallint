// Package plugin is the entry point for golangci-lint's v2 module-plugin
// system. Configure golangci-lint with a "plugins" entry pointing at this
// package and it will load every analyzer in the all/ registry.
//
// Example .custom-gcl.yml:
//
//	version: v2.0.0
//	plugins:
//	  - module: github.com/mattconzen/temporallint
//	    path: ./tools/temporallint
package plugin

import (
	"golang.org/x/tools/go/analysis"

	"github.com/mattconzen/temporallint/all"
)

// New is the well-known entry point golangci-lint expects for module
// plugins. The settings argument is forwarded by the host but unused —
// every rule is configured via its own analyzer flags.
func New(_ any) ([]*analysis.Analyzer, error) {
	return all.Analyzers(), nil
}
