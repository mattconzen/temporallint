package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.ActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second, // want `is shorter than RetryPolicy.InitialInterval`
		RetryPolicy: &workflow.RetryPolicyAlias{
			InitialInterval: 10 * time.Second,
		},
	}
	return nil
}
