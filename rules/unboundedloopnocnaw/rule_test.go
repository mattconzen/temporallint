package unboundedloopnocnaw_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/unboundedloopnocnaw"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, unboundedloopnocnaw.Analyzer, "violation")
	analysistest.Run(t, dir, unboundedloopnocnaw.Analyzer, "clean")
}
