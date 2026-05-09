package localactivitymisuse_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/localactivitymisuse"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, localactivitymisuse.Analyzer, "violation")
	analysistest.Run(t, dir, localactivitymisuse.Analyzer, "clean")
}
