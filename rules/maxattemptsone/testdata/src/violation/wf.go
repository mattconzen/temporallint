package violation

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	policy := temporal.RetryPolicy{
		InitialInterval: time.Second,
		MaximumAttempts: 1, // want `MaximumAttempts: 1 silently disables retry`
	}
	_ = policy
	return nil
}
