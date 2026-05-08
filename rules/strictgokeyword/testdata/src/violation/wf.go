package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	go func() { // want `bare .go. statement in workflow code`
		_ = 1
	}()
	return nil
}
