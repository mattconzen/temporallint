package violation

import (
	"context"
	"errors"

	_ "go.temporal.io/sdk/activity"
)

type Result struct{ Value int }

func MyActivity(ctx context.Context, x int) (Result, error) {
	if x > 0 {
		return Result{Value: x}, errors.New("partial success") // want `both a non-zero payload and a non-nil error`
	}
	return Result{}, nil
}
