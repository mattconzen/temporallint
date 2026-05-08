package clean

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.Now(ctx)
	return nil
}

// time.Now is fine outside workflow code.
func notAWorkflow() time.Time { return time.Now() }
