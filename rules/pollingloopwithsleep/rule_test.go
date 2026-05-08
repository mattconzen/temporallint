package pollingloopwithsleep_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/pollingloopwithsleep"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, pollingloopwithsleep.Analyzer, "violation")
	analysistest.Run(t, dir, pollingloopwithsleep.Analyzer, "clean")
}
