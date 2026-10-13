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
	ListWorktrees(ctx context.Context, repoDir string) (string, error)
	Status(ctx context.Context, wsDir string) (string, error)
	AheadBehind(ctx context.Context, wsDir string) (int, int, error)
	DiffCached(ctx context.Context, wsDir string) (string, error)
	Diff(ctx context.Context, wsDir string) (string, error)
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

// ListWorktrees executes `git worktree list --porcelain -z`.
func (r *NativeRunner) ListWorktrees(ctx context.Context, repoDir string) (string, error) {
	return r.Exec(ctx, repoDir, "worktree", "list", "--porcelain", "-z")
}

// Status executes `git status --porcelain -z`.
func (r *NativeRunner) Status(ctx context.Context, wsDir string) (string, error) {
	return r.Exec(ctx, wsDir, "status", "--porcelain", "-z")
}

// AheadBehind queries commits ahead and behind upstream. Returns (0, 0, nil) if no upstream is set.
func (r *NativeRunner) AheadBehind(ctx context.Context, wsDir string) (int, int, error) {
	out, err := r.Exec(ctx, wsDir, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		// Typically means no upstream configured, which is normal for local branches.
		return 0, 0, nil
	}

	fields := strings.Fields(strings.TrimSpace(out))
	if len(fields) != 2 {
		return 0, 0, nil
	}

	var ahead, behind int
	fmt.Sscanf(fields[0], "%d", &ahead)
	fmt.Sscanf(fields[1], "%d", &behind)
	return ahead, behind, nil
}

// DiffCached returns staged diff.
func (r *NativeRunner) DiffCached(ctx context.Context, wsDir string) (string, error) {
	return r.Exec(ctx, wsDir, "diff", "--cached")
}

// Diff returns unstaged diff.
func (r *NativeRunner) Diff(ctx context.Context, wsDir string) (string, error) {
	return r.Exec(ctx, wsDir, "diff")
}
