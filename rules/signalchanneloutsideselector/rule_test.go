package signalchanneloutsideselector_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/signalchanneloutsideselector"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, signalchanneloutsideselector.Analyzer, "violation")
	analysistest.Run(t, dir, signalchanneloutsideselector.Analyzer, "clean")
}
