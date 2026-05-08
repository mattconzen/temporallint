package clean

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func LongActivity(ctx context.Context) error {
	activity.RecordHeartbeat(ctx, "progress")
	return nil
}

func WF(ctx workflow.Context) error {
	_ = workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    30 * time.Second,
	}
	return nil
}
