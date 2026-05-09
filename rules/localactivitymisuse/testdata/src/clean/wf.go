package clean

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, opts)
	_ = workflow.ExecuteActivity(ctx, "act")
	return nil
}
