package clean

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	t := workflow.NewTimer(ctx, time.Second)
	return t.Get(ctx, nil)
}
