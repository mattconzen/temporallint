// Package tooshorttimeouts flags ActivityOptions whose
// StartToCloseTimeout (or ScheduleToCloseTimeout) is shorter than a
// configurable threshold (default 1s). A too-short timeout makes
// transient downstream latency look like a hung activity, triggering
// false retries and noise. Adjust the threshold with
// `-tooshorttimeouts.threshold=<duration>`.
package tooshorttimeouts

import (
	"fmt"
	"go/ast"
	"time"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mattconzen/temporallint/temporalctx"
)

var threshold = time.Second

var Analyzer = &analysis.Analyzer{
	Name:     "tooshorttimeouts",
	Doc:      "ActivityOptions StartToCloseTimeout / ScheduleToCloseTimeout shorter than the threshold (default 1s) causes false retries.",
	URL:      "https://github.com/jlegrone/100-temporal-mistakes#setting-too-short-timeouts",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func init() {
	Analyzer.Flags.DurationVar(&threshold, "threshold", time.Second,
		"minimum acceptable activity timeout (default 1s)")
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		cl := n.(*ast.CompositeLit)
		if !temporalctx.IsActivityOptions(pass, cl) {
			return
		}
		for _, name := range []string{"StartToCloseTimeout", "ScheduleToCloseTimeout"} {
			kv := temporalctx.FieldByName(cl, name)
			if kv == nil {
				continue
			}
			d, ok := temporalctx.DurationLiteralValue(pass, kv.Value)
			if !ok {
				continue
			}
			if d <= 0 || d >= threshold {
				continue
			}
			pass.Report(analysis.Diagnostic{
				Pos: kv.Pos(), End: kv.End(),
				Message: fmt.Sprintf("%s (%s) is shorter than the %s threshold; transient latency will trigger false retries", name, d, threshold),
			})
		}
	})
	return nil, nil
}
