package main

import (
	"go.temporal.io/sdk/worker"
)

func main() {
	_ = worker.New(nil, "orders-tq", worker.Options{})
}
