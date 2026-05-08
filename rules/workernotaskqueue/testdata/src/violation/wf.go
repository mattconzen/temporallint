package main

import (
	"go.temporal.io/sdk/worker"
)

func main() {
	_ = worker.New(nil, "", worker.Options{}) // want `empty task queue`
}
