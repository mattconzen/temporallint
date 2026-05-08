package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	<-time.After(time.Second) // want `time.After in workflow code is non-deterministic`
	return nil
}
