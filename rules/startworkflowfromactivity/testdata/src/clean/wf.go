package clean

import (
	"context"

	_ "go.temporal.io/sdk/activity"
)

func MyActivity(ctx context.Context) error {
	// Activity does its own work; orchestration lives in the workflow.
	return nil
}
