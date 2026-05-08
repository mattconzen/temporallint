package unboundednoceiling_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/unboundednoceiling"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, unboundednoceiling.Analyzer, "violation")
	analysistest.Run(t, dir, unboundednoceiling.Analyzer, "clean")
}
