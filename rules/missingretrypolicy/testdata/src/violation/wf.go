package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	opts := workflow.ActivityOptions{ // want `no RetryPolicy`
		StartToCloseTimeout: time.Minute,
	}
	_ = workflow.WithActivityOptions(ctx, opts)
	return nil
}
