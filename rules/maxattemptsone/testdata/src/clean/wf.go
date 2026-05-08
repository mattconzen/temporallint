package clean

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = temporal.RetryPolicy{
		InitialInterval: time.Second,
		MaximumAttempts: 5,
	}
	_ = temporal.RetryPolicy{
		InitialInterval: time.Second,
		// MaximumAttempts unset = unlimited (default), not 1
	}
	_ = workflow.Now(ctx)
	return nil
}
