package clean

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &workflow.RetryPolicyAlias{
			InitialInterval: time.Second,
			MaximumAttempts: 5,
		},
	}
	_ = workflow.WithActivityOptions(ctx, opts)
	_ = temporal.RetryPolicy{} // ensures stub is exercised
	return nil
}
