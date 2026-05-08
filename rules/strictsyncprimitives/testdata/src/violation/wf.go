package violation

import (
	"sync"

	"go.temporal.io/sdk/workflow"
)

func WF(ctx workflow.Context) error {
	var mu sync.Mutex // want `sync.Mutex in workflow code`
	mu.Lock()
	mu.Unlock()
	var wg sync.WaitGroup // want `sync.WaitGroup in workflow code`
	_ = wg
	return nil
}
