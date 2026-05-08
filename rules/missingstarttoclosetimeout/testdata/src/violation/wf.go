package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func MissingBoth(ctx workflow.Context) error {
	opts := workflow.ActivityOptions{ // want `must set StartToCloseTimeout`
		HeartbeatTimeout: time.Minute,
	}
	_ = workflow.WithActivityOptions(ctx, opts)
	return nil
}

func InlineLiteral(ctx workflow.Context) error {
	_ = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{ // want `must set StartToCloseTimeout`
		TaskQueue: "tq",
	})
	return nil
}
