package workernotaskqueue_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/monorepo/tools/temporallint/rules/workernotaskqueue"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, workernotaskqueue.Analyzer, "violation")
	analysistest.Run(t, dir, workernotaskqueue.Analyzer, "clean")
}
