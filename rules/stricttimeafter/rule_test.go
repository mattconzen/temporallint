package stricttimeafter_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/stricttimeafter"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, stricttimeafter.Analyzer, "violation")
	analysistest.Run(t, dir, stricttimeafter.Analyzer, "clean")
}
