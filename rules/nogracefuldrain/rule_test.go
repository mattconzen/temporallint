package nogracefuldrain_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/nogracefuldrain"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, nogracefuldrain.Analyzer, "violation")
	analysistest.Run(t, dir, nogracefuldrain.Analyzer, "clean")
}
