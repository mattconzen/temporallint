package strictsyncprimitives_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictsyncprimitives"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, strictsyncprimitives.Analyzer, "violation")
	analysistest.Run(t, dir, strictsyncprimitives.Analyzer, "clean")
}
