package payloadanderror_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/payloadanderror"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, payloadanderror.Analyzer, "violation")
	analysistest.Run(t, dir, payloadanderror.Analyzer, "clean")
}
