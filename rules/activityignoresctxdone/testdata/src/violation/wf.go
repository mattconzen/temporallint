package violation

import (
	"context"

	_ "go.temporal.io/sdk/activity"
)

func MyActivity(ctx context.Context) error {
	for i := 0; i < 100; i++ { // want `cannot be cancelled promptly`
		_ = i
	}
	return nil
}
