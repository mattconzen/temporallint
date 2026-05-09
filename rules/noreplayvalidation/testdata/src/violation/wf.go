package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error { // want `package contains workflow definitions but no replay test`
	return nil
}
