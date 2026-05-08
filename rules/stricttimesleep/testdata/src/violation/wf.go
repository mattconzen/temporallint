package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	time.Sleep(time.Second) // want `time.Sleep blocks the worker thread`
	return nil
}
