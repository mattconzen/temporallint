package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.ActivityOptions{
		StartToCloseTimeout: 50 * time.Millisecond, // want `StartToCloseTimeout .* shorter than the 1s threshold`
	}
	_ = workflow.ActivityOptions{
		ScheduleToCloseTimeout: 100 * time.Millisecond, // want `ScheduleToCloseTimeout .* shorter than the 1s threshold`
	}
	return nil
}
