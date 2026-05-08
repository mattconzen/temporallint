package violation

import (
	"log"
	"os"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	if false {
		os.Exit(1) // want `os.Exit in workflow code`
	}
	log.Fatal("nope") // want `log.Fatal in workflow code`
	return nil
}
