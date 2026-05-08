package clean

import (
	"context"

	"go.temporal.io/sdk/client"
)

func DoIt(ctx context.Context, c client.Client) error {
	return c.CancelWorkflow(ctx, "wf-1", "")
}
