package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	for i := 0; i < 100; i++ { // want `polling loop in workflow uses workflow.Sleep`
		_ = workflow.Now(ctx)
		_ = workflow.Sleep(ctx, time.Second)
	}
	return nil
}
