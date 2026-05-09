package strictgokeyword_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/strictgokeyword"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, strictgokeyword.Analyzer, "violation")
	analysistest.Run(t, dir, strictgokeyword.Analyzer, "clean")
}
