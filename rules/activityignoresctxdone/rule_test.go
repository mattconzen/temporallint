package activityignoresctxdone_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/activityignoresctxdone"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, activityignoresctxdone.Analyzer, "violation")
	analysistest.Run(t, dir, activityignoresctxdone.Analyzer, "clean")
}
