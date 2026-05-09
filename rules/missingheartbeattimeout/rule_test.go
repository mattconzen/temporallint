package missingheartbeattimeout_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/missingheartbeattimeout"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, missingheartbeattimeout.Analyzer, "violation")
	analysistest.Run(t, dir, missingheartbeattimeout.Analyzer, "clean")
}
