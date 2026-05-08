package violation

import (
	"context"

	"go.temporal.io/sdk/client"
)

func DoIt(ctx context.Context, c client.Client) error {
	return c.TerminateWorkflow(ctx, "wf-1", "", "shutdown") // want `TerminateWorkflow skips graceful cancellation cleanup`
}
