package clean

import (
	"context"

	"go.temporal.io/sdk/activity"
)

func MyActivity(ctx context.Context) error {
	for i := 0; i < 100; i++ {
		activity.RecordHeartbeat(ctx, i)
	}
	return nil
}
