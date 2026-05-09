package clean

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	}
	_ = workflow.ActivityOptions{
		ScheduleToCloseTimeout: 30 * time.Second,
	}
	return nil
}
