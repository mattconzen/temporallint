package violation

import (
	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	for { // want `unbounded for\{\} loop`
		_ = workflow.Now(ctx)
	}
}
