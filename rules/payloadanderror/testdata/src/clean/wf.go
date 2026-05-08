package clean

import (
	"context"
	"errors"

	_ "go.temporal.io/sdk/activity"
)

type Result struct{ Value int }

func MyActivity(ctx context.Context, x int) (Result, error) {
	if x < 0 {
		return Result{}, errors.New("bad input") // zero result + error: clean
	}
	return Result{Value: x}, nil // result + nil error: clean
}
