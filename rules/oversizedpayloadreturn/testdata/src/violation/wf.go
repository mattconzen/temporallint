package violation

import (
	"context"

	_ "go.temporal.io/sdk/activity"
)

type FatResult struct {
	A1, A2, A3, A4, A5, A6, A7, A8 string
	B1, B2, B3, B4, B5, B6, B7, B8 int
	C1, C2                         []byte
}

func MyActivity(ctx context.Context) (FatResult, error) { // want `has 18 fields`
	return FatResult{}, nil
}
