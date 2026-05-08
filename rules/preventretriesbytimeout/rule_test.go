package preventretriesbytimeout_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/preventretriesbytimeout"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, preventretriesbytimeout.Analyzer, "violation")
	analysistest.Run(t, dir, preventretriesbytimeout.Analyzer, "clean")
}
