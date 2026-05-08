package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	workflow.Await(ctx) // want `Await without NewTimer / AwaitWithTimeout`
	return nil
}
