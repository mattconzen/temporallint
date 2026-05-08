package clean

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	state := 0
	_ = workflow.SetQueryHandler(ctx, "state", func() (int, error) {
		return state, nil
	})
	return nil
}
