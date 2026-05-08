package strictmakechan_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/strictmakechan"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, strictmakechan.Analyzer, "violation")
	analysistest.Run(t, dir, strictmakechan.Analyzer, "clean")
}
