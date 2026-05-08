package clean

import (
	"go.temporal.io/sdk/workflow"
)

// Workflows must NOT call math/rand directly. To capture randomness
// deterministically, wrap an activity (or use SideEffect) and pass the
// result back. The linter currently flags math/rand calls anywhere inside
// the workflow function body — including inside workflow.SideEffect — so
// the recommended pattern is to keep rand entirely outside workflow code.
func WF(ctx workflow.Context) error {
	var n int
	_ = workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		// rand would go here in production code — see roadmap for
		// allow-list of math/rand inside SideEffect callbacks.
		return 7
	})
	_ = n
	return nil
}
