package noheartbeatdetails_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/noheartbeatdetails"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, noheartbeatdetails.Analyzer, "violation")
	analysistest.Run(t, dir, noheartbeatdetails.Analyzer, "clean")
}
