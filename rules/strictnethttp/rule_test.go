package strictnethttp_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictnethttp"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, strictnethttp.Analyzer, "violation")
	analysistest.Run(t, dir, strictnethttp.Analyzer, "clean")
}
