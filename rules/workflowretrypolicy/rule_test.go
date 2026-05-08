package workflowretrypolicy_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/workflowretrypolicy"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, workflowretrypolicy.Analyzer, "violation")
	analysistest.Run(t, dir, workflowretrypolicy.Analyzer, "clean")
}
