package startworkflowbadtaskqueue_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/startworkflowbadtaskqueue"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, startworkflowbadtaskqueue.Analyzer, "violation")
	analysistest.Run(t, dir, startworkflowbadtaskqueue.Analyzer, "clean")
}
