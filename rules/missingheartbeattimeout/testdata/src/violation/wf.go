package violation

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
	_ = workflow.ActivityOptions{ // want `package contains activity.RecordHeartbeat calls but this ActivityOptions has no HeartbeatTimeout`
		StartToCloseTimeout: 5 * time.Minute,
	}
	return nil
}
