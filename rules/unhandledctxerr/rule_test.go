package unhandledctxerr_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/unhandledctxerr"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, unhandledctxerr.Analyzer, "violation")
	analysistest.Run(t, dir, unhandledctxerr.Analyzer, "clean")
}
