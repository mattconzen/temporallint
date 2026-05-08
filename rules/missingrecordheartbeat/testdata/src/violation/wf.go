package violation

import (
	"context"
	"time"

	_ "go.temporal.io/sdk/activity"
)

func MyActivity(ctx context.Context) error {
	for i := 0; i < 100; i++ { // want `long-running activity loop without activity.RecordHeartbeat`
		time.Sleep(time.Second)
		_ = i
	}
	return nil
}
