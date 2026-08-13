package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Docker runs a command inside a container with workspace mounted at /workspace.
type Docker struct {
	Policy Policy
}

// NewDocker creates a Docker sandbox.
func NewDocker(p Policy) *Docker {
	if p.DockerBin == "" {
		p = DefaultPolicy()
	}
	return &Docker{Policy: p}
}

// Available reports whether docker CLI works.
func (d *Docker) Available(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, d.Policy.DockerBin, "version", "--format", "{{.Server.Version}}")
	return cmd.Run() == nil
}

// Run executes command in the sandbox. workspace is mounted at /workspace as cwd.
func (d *Docker) Run(ctx context.Context, workspace, command string) (stdout string, err error) {
	timeout := d.Policy.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := d.Policy.RunArgs(workspace, command)

	cmd := exec.CommandContext(ctx, d.Policy.DockerBin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	out := outBuf.String()
	if errBuf.Len() > 0 {
		out = strings.TrimSpace(out + "\n" + errBuf.String())
	}
	if ctx.Err() == context.DeadlineExceeded {
		return out, fmt.Errorf("sandbox timeout after %s", timeout)
	}
	if runErr != nil {
		return out, fmt.Errorf("sandbox: %w", runErr)
	}
	return out, nil
}
