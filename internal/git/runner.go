package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner defines the interface for interacting with the native Git binary.
type Runner interface {
	Exec(ctx context.Context, dir string, args ...string) (string, error)
	Version(ctx context.Context) (string, error)
}

// NativeRunner executes Git via system subprocesses.
type NativeRunner struct {
	gitBinPath string
}

// NewNativeRunner creates a new NativeRunner looking for the 'git' binary in PATH.
func NewNativeRunner() (*NativeRunner, error) {
	path, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git binary not found in system PATH: %w", err)
	}
	return &NativeRunner{gitBinPath: path}, nil
}

// Exec executes a git command in the specified directory with structured arguments.
func (r *NativeRunner) Exec(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, r.gitBinPath, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errStr := strings.TrimSpace(stderr.String())
		if errStr == "" {
			errStr = strings.TrimSpace(stdout.String())
		}
		if errStr == "" {
			errStr = err.Error()
		}
		return "", fmt.Errorf("git error [%v]: %s", args, errStr)
	}

	return stdout.String(), nil
}

// Version returns the output of `git --version`.
func (r *NativeRunner) Version(ctx context.Context) (string, error) {
	out, err := r.Exec(ctx, "", "--version")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
