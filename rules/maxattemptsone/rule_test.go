package maxattemptsone_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/maxattemptsone"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, maxattemptsone.Analyzer, "violation")
	analysistest.Run(t, dir, maxattemptsone.Analyzer, "clean")
}
