package missingdisconnectedcontextcleanup_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingdisconnectedcontextcleanup"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, missingdisconnectedcontextcleanup.Analyzer, "violation")
	analysistest.Run(t, dir, missingdisconnectedcontextcleanup.Analyzer, "clean")
}
