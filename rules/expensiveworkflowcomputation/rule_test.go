package expensiveworkflowcomputation_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/expensiveworkflowcomputation"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, expensiveworkflowcomputation.Analyzer, "violation")
	analysistest.Run(t, dir, expensiveworkflowcomputation.Analyzer, "clean")
}
