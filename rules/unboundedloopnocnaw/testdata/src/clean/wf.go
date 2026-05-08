package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	for i := 0; i < 100; i++ {
		_ = workflow.Now(ctx)
	}
	// Package contains a ContinueAsNew call → unbounded loops are allowed.
	return workflow.ContinueAsNew(ctx, WF)
}

func Other(ctx workflow.Context) error {
	for {
		_ = workflow.Now(ctx)
	}
}
