package clean

import (
	"errors"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	return errors.New("graceful failure")
}
