package oversizedpayloadreturn_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/oversizedpayloadreturn"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, oversizedpayloadreturn.Analyzer, "violation")
	analysistest.Run(t, dir, oversizedpayloadreturn.Analyzer, "clean")
}
