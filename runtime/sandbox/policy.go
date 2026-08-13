package sandbox

import "time"

// Policy controls Docker sandbox behavior.
type Policy struct {
	Image         string
	Network       string
	Memory        string
	CPUs          string
	WorkspaceMode string // rw | ro
	Timeout       time.Duration
	DockerBin     string
	User          string
	ReadOnlyRoot  bool
	PidsLimit     string
	DropCaps      bool
	NoNewPrivs    bool
}

// DefaultPolicy returns a restrictive MVP policy.
func DefaultPolicy() Policy {
	return Policy{
		Image:         "alpine:3.20",
		Network:       "none",
		Memory:        "256m",
		CPUs:          "1",
		WorkspaceMode: "rw",
		Timeout:       30 * time.Second,
		DockerBin:     "docker",
		DropCaps:      true,
		NoNewPrivs:    true,
		PidsLimit:     "64",
	}
}

// ProductionPolicy is fail-closed: no network, non-root, cap drop, read-only root.
func ProductionPolicy() Policy {
	p := DefaultPolicy()
	p.User = "65534:65534"
	p.ReadOnlyRoot = true
	p.Network = "none"
	p.DropCaps = true
	p.NoNewPrivs = true
	p.PidsLimit = "64"
	return p
}

// RunArgs builds `docker run` arguments (without the docker binary).
func (p Policy) RunArgs(workspace, command string) []string {
	if p.Network == "" {
		p.Network = "none"
	}
	if p.WorkspaceMode == "" {
		p.WorkspaceMode = "rw"
	}
	if p.Memory == "" {
		p.Memory = "256m"
	}
	if p.CPUs == "" {
		p.CPUs = "1"
	}
	if p.Image == "" {
		p.Image = "alpine:3.20"
	}
	mount := workspace + ":/workspace:" + p.WorkspaceMode
	args := []string{
		"run", "--rm",
		"--network", p.Network,
		"--memory", p.Memory,
		"--cpus", p.CPUs,
		"--workdir", "/workspace",
		"-v", mount,
	}
	if p.User != "" {
		args = append(args, "--user", p.User)
	}
	if p.ReadOnlyRoot {
		args = append(args, "--read-only", "--tmpfs", "/tmp:rw,noexec,nosuid,size=64m")
	}
	if p.DropCaps {
		args = append(args, "--cap-drop", "ALL")
	}
	if p.NoNewPrivs {
		args = append(args, "--security-opt", "no-new-privileges")
	}
	if p.PidsLimit != "" {
		args = append(args, "--pids-limit", p.PidsLimit)
	}
	args = append(args, p.Image, "sh", "-c", command)
	return args
}
