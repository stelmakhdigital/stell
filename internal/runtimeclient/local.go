package runtimeclient

import (
	"context"

	"github.com/budaev/stell/runtime/executor"
)

// LocalClient executes tools in-process via Hands executor (MVP convenience).
type LocalClient struct {
	Exec *executor.Executor
}

// NewLocal wraps an executor.
func NewLocal(exec *executor.Executor) *LocalClient {
	return &LocalClient{Exec: exec}
}

// Execute delegates to the executor.
func (c *LocalClient) Execute(ctx context.Context, req ExecuteRequest) (ExecuteResponse, error) {
	res, err := c.Exec.Execute(ctx, executor.Request{
		Tool:      req.Tool,
		Args:      req.Args,
		Workspace: req.Workspace,
		TimeoutMs: req.TimeoutMs,
	})
	if err != nil {
		return ExecuteResponse{Error: err.Error()}, nil
	}
	return ExecuteResponse{
		Content:   res.Content,
		Truncated: res.Truncated,
		Error:     res.Error,
		Metadata:  res.Metadata,
	}, nil
}
