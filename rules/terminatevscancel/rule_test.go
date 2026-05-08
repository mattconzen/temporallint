package terminatevscancel_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/terminatevscancel"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, terminatevscancel.Analyzer, "violation")
	analysistest.Run(t, dir, terminatevscancel.Analyzer, "clean")
}
