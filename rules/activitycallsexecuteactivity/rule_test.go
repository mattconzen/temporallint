package activitycallsexecuteactivity_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/activitycallsexecuteactivity"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, activitycallsexecuteactivity.Analyzer, "violation")
	analysistest.Run(t, dir, activitycallsexecuteactivity.Analyzer, "clean")
}
