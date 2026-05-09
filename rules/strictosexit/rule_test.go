package strictosexit_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/strictosexit"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, strictosexit.Analyzer, "violation")
	analysistest.Run(t, dir, strictosexit.Analyzer, "clean")
}
