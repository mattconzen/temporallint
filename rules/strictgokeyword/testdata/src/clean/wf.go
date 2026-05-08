package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	workflow.Go(ctx, func(workflow.Context) {
		_ = 1
	})
	return nil
}
