package tooshorttimeouts_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/tooshorttimeouts"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, tooshorttimeouts.Analyzer, "violation")
	analysistest.Run(t, dir, tooshorttimeouts.Analyzer, "clean")
}
