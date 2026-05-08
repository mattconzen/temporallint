package violation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = time.Now() // want `time.Now\(\) is non-deterministic`
	return nil
}
