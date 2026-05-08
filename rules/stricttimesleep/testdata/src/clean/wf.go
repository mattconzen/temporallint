package clean

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	return workflow.Sleep(ctx, time.Second)
}
