package missingrecordheartbeat_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/missingrecordheartbeat"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, missingrecordheartbeat.Analyzer, "violation")
	analysistest.Run(t, dir, missingrecordheartbeat.Analyzer, "clean")
}
