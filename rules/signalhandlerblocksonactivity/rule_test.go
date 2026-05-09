package signalhandlerblocksonactivity_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/signalhandlerblocksonactivity"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, signalhandlerblocksonactivity.Analyzer, "violation")
	analysistest.Run(t, dir, signalhandlerblocksonactivity.Analyzer, "clean")
}
