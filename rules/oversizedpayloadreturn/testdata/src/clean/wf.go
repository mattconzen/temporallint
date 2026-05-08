package clean

import (
	"context"

	_ "go.temporal.io/sdk/activity"
)

type SmallResult struct {
	ID    string
	Value int
}

func MyActivity(ctx context.Context) (SmallResult, error) {
	return SmallResult{}, nil
}
