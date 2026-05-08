package clean

import (
	"context"

	"go.temporal.io/sdk/activity"
)

func WithDone(ctx context.Context) error {
	for i := 0; i < 100; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_ = i
	}
	return nil
}

func WithHeartbeat(ctx context.Context) error {
	for i := 0; i < 100; i++ {
		activity.RecordHeartbeat(ctx, i)
	}
	return nil
}
