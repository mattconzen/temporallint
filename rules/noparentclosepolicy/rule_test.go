package noparentclosepolicy_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/noparentclosepolicy"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, noparentclosepolicy.Analyzer, "violation")
	analysistest.Run(t, dir, noparentclosepolicy.Analyzer, "clean")
}
