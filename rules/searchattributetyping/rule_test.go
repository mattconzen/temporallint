package searchattributetyping_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/searchattributetyping"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, searchattributetyping.Analyzer, "violation")
	analysistest.Run(t, dir, searchattributetyping.Analyzer, "clean")
}
