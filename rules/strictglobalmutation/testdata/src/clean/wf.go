package clean

import (
	"go.temporal.io/sdk/workflow"
)

type State struct{ Counter int }

func WF(ctx workflow.Context, s State) error {
	s.Counter++ // local copy
	_ = s
	return nil
}
