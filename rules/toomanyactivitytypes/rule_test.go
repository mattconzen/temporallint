package toomanyactivitytypes_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/toomanyactivitytypes"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, toomanyactivitytypes.Analyzer, "violation")
	analysistest.Run(t, dir, toomanyactivitytypes.Analyzer, "clean")
}
