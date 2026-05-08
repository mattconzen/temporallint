package queryhandlerwithsideeffects_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/queryhandlerwithsideeffects"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, queryhandlerwithsideeffects.Analyzer, "violation")
	analysistest.Run(t, dir, queryhandlerwithsideeffects.Analyzer, "clean")
}
