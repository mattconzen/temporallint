package violation

import (
	"os"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_ = os.Getenv("FOO")    // want `os.Getenv in workflow code`
	_, _ = os.LookupEnv("X") // want `os.LookupEnv in workflow code`
	_ = os.Environ()         // want `os.Environ in workflow code`
	return nil
}
