package signaldrainmissing_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/signaldrainmissing"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, signaldrainmissing.Analyzer, "violation")
	analysistest.Run(t, dir, signaldrainmissing.Analyzer, "clean")
}
