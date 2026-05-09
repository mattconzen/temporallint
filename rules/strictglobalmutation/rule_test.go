package strictglobalmutation_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/strictglobalmutation"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, strictglobalmutation.Analyzer, "violation")
	analysistest.Run(t, dir, strictglobalmutation.Analyzer, "clean")
}
