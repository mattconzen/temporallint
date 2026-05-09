package workflowidreusepolicymismatch_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/workflowidreusepolicymismatch"
)

func TestAnalyzer(t *testing.T) {
	dir := analysistest.TestData()
	analysistest.Run(t, dir, workflowidreusepolicymismatch.Analyzer, "violation")
	analysistest.Run(t, dir, workflowidreusepolicymismatch.Analyzer, "clean")
}
