package clean

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &workflow.RetryPolicyAlias{
			InitialInterval: 10 * time.Second,
		},
	}
	return nil
}
