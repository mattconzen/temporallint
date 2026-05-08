package stricttimenow_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/stricttimenow"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, stricttimenow.Analyzer, "violation")
	analysistest.Run(t, dir, stricttimenow.Analyzer, "clean")
}
