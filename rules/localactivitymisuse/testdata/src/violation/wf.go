package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	opts := workflow.LocalActivityOptions{ // want `LocalActivityOptions in use`
		StartToCloseTimeout: time.Second,
	}
	ctx = workflow.WithLocalActivityOptions(ctx, opts)
	_ = workflow.ExecuteLocalActivity(ctx, "act") // want `ExecuteLocalActivity in use`
	return nil
}
