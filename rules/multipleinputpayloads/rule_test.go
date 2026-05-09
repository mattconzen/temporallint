package multipleinputpayloads_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/multipleinputpayloads"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, multipleinputpayloads.Analyzer, "violation")
	analysistest.Run(t, dir, multipleinputpayloads.Analyzer, "clean")
}
