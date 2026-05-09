package strictosgetenv_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/strictosgetenv"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, strictosgetenv.Analyzer, "violation")
	analysistest.Run(t, dir, strictosgetenv.Analyzer, "clean")
}
