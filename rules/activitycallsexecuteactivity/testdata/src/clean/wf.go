package clean

import (
	"context"

	_ "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func MyActivity(ctx context.Context) error {
	// Activity does its own work — no nested ExecuteActivity.
	return nil
}

func WF(ctx workflow.Context) error {
	// Workflows MAY call ExecuteActivity — that's fine.
	_ = workflow.ExecuteActivity(ctx, MyActivity)
	return nil
}
