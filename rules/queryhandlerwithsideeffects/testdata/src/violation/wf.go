package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = workflow.SetQueryHandler(ctx, "state", func() (int, error) {
		_ = workflow.Sleep(ctx, time.Second) // want `query handler calls workflow.Sleep`
		return 0, nil
	})
	return nil
}
