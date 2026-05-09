package noreplayvalidation_test

import (
	"flag"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mattconzen/temporallint/rules/noreplayvalidation"
)

func TestAnalyzer(t *testing.T) {
	// The rule is default-off. Enable it via the analyzer's own flag for
	// the duration of the test.
	if err := noreplayvalidation.Analyzer.Flags.Set("noreplayvalidation", "true"); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	t.Cleanup(func() {
		_ = noreplayvalidation.Analyzer.Flags.Set("noreplayvalidation", "false")
	})

	dir := analysistest.TestData()
	analysistest.Run(t, dir, noreplayvalidation.Analyzer, "violation")
	analysistest.Run(t, dir, noreplayvalidation.Analyzer, "clean")

	// Touch flag pkg so the import is "used" if we ever drop the inner
	// flag.Set above.
	_ = flag.CommandLine
}
