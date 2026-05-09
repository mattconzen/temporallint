package sideeffectnoresult_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/sideeffectnoresult"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, sideeffectnoresult.Analyzer, "violation")
	analysistest.Run(t, dir, sideeffectnoresult.Analyzer, "clean")
}
