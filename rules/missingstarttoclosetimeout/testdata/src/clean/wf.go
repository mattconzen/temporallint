package clean

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WithStartToClose(ctx workflow.Context) error {
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	_ = workflow.WithActivityOptions(ctx, opts)
	return nil
}

func WithScheduleToClose(ctx workflow.Context) error {
	opts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:       time.Minute,
	}
	_ = workflow.WithActivityOptions(ctx, opts)
	return nil
}
