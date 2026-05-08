package violation

import (
	"context"

	_ "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
)

func MyActivity(ctx context.Context, c client.Client) error {
	_, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{}, nil) // want `activity should not call client.ExecuteWorkflow`
	return err
}
