package startworkflowfromactivity_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/startworkflowfromactivity"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, startworkflowfromactivity.Analyzer, "violation")
	analysistest.Run(t, dir, startworkflowfromactivity.Analyzer, "clean")
}
