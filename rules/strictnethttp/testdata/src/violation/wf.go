package violation

import (
	"net/http"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	_, _ = http.Get("https://example.com") // want `net/http call in workflow code`
	return nil
}
