package clean

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
)

func MyActivity(ctx context.Context) error {
	for i := 0; i < 100; i++ {
		time.Sleep(time.Second)
		activity.RecordHeartbeat(ctx, i)
	}
	return nil
}

// Short loop is fine without heartbeat.
func Short(ctx context.Context) error {
	for i := 0; i < 3; i++ {
		_ = i
	}
	return nil
}
