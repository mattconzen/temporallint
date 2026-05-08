package strictmathrand_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictmathrand"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, strictmathrand.Analyzer, "violation")
	analysistest.Run(t, dir, strictmathrand.Analyzer, "clean")
}
