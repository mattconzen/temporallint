package missingworkflowtimeout_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/missingworkflowtimeout"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, missingworkflowtimeout.Analyzer, "violation")
	analysistest.Run(t, dir, missingworkflowtimeout.Analyzer, "clean")
}
