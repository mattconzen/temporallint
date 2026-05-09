package awaitnotimeout_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/awaitnotimeout"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, awaitnotimeout.Analyzer, "violation")
	analysistest.Run(t, dir, awaitnotimeout.Analyzer, "clean")
}
