package main

import (
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

var _ = worker.New(nil, "orders-tq", worker.Options{})

var _ = client.StartWorkflowOptions{
	ID:                       "wf-1",
	TaskQueue:                "orders-tq",
	WorkflowExecutionTimeout: time.Hour,
}
