package missingretrypolicy_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/missingretrypolicy"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, missingretrypolicy.Analyzer, "violation")
	analysistest.Run(t, dir, missingretrypolicy.Analyzer, "clean")
}
