package violation

import (
	"go.temporal.io/sdk/workflow"
)

var counter int

func WF(ctx workflow.Context) error {
	counter = counter + 1 // want `writing to a package-level variable from workflow code`
	return nil
}
