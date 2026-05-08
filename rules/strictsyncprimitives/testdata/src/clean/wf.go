package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	// Workflow scheduler is single-threaded — no mutex needed.
	state := 0
	state++
	_ = state
	return nil
}
