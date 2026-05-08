package clean

import (
	"os"

	"go.temporal.io/sdk/workflow"
)

type Input struct{ Foo string }

func WF(ctx workflow.Context, in Input) error {
	_ = in.Foo
	return nil
}

func notAWorkflow() string { return os.Getenv("FOO") }
