package clean

import (
	"net/http"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	// I/O lives in an activity, not the workflow.
	return nil
}

// Activities are NOT workflow code — http calls are fine here.
func httpActivity() (*http.Response, error) {
	return http.Get("https://example.com")
}
