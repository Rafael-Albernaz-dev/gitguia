package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Rafael-Albernaz-dev/gitguia/internal/domain"
	"github.com/Rafael-Albernaz-dev/gitguia/internal/git"
)

// Service defines the core use case contracts available to both CLI and TUI.
type Service interface {
	GetGitVersion(ctx context.Context) (string, error)
	ResolveProject(ctx context.Context, dir string) (*domain.Project, error)
	ListWorktrees(ctx context.Context, currentDir string) ([]*domain.Workspace, error)
	GetWorkspaceStatus(ctx context.Context, currentDir string, targetIdentifier string) (*domain.Workspace, *domain.ChangeSet, error)
}

// App implements the Service interface coordinating adapters and business logic.
type App struct {
	gitRunner git.Runner
}

// NewApp initializes the application core with required adapters.
func NewApp(gitRunner git.Runner) *App {
	return &App{
		gitRunner: gitRunner,
	}
}

// GetGitVersion queries the host git binary version.
func (a *App) GetGitVersion(ctx context.Context) (string, error) {
	return a.gitRunner.Version(ctx)
}

// ResolveProject inspects the given directory and discovers the project Git root and common dir.
func (a *App) ResolveProject(ctx context.Context, dir string) (*domain.Project, error) {
	commonDir, err := a.gitRunner.Exec(ctx, dir, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, fmt.Errorf("not a git repository (or any parent up to mount point): %w", err)
	}

	rootDir, err := a.gitRunner.Exec(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("unable to resolve top-level git repository: %w", err)
	}

	commonDir = strings.TrimSpace(commonDir)
	rootDir = strings.TrimSpace(rootDir)

	return &domain.Project{
		ID:                commonDir,
		RootPath:          rootDir,
		GitDir:            commonDir,
		DefaultBaseBranch: "main",
	}, nil
}

// ListWorktrees discovers all worktrees and concurrently inspects their status.
func (a *App) ListWorktrees(ctx context.Context, currentDir string) ([]*domain.Workspace, error) {
	proj, err := a.ResolveProject(ctx, currentDir)
	if err != nil {
		return nil, err
	}

	raw, err := a.gitRunner.ListWorktrees(ctx, proj.RootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	worktrees, err := git.ParseWorktreeList(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse worktrees: %w", err)
	}

	// Resolve current absolute path
	absCurrentDir, _ := filepath.Abs(currentDir)

	// Concurrently query status for each worktree (worker pool / sync.WaitGroup)
	var wg sync.WaitGroup
	for _, ws := range worktrees {
		absWsPath, _ := filepath.Abs(ws.Path)
		if absWsPath == absCurrentDir {
			ws.IsCurrent = true
		}

		wg.Add(1)
		go func(w *domain.Workspace) {
			defer wg.Done()
			statusRaw, err := a.gitRunner.Status(ctx, w.Path)
			if err != nil {
				return
			}
			cs, err := git.ParseStatus(statusRaw, w.Path, w.Branch)
			if err != nil {
				return
			}
			ahead, behind, _ := a.gitRunner.AheadBehind(ctx, w.Path)
			cs.Ahead = ahead
			cs.Behind = behind
			w.Status = cs
		}(ws)
	}
	wg.Wait()

	return worktrees, nil
}

// GetWorkspaceStatus fetches detailed changes and diffs for a specific workspace.
func (a *App) GetWorkspaceStatus(ctx context.Context, currentDir string, targetIdentifier string) (*domain.Workspace, *domain.ChangeSet, error) {
	worktrees, err := a.ListWorktrees(ctx, currentDir)
	if err != nil {
		return nil, nil, err
	}

	var target *domain.Workspace
	if targetIdentifier == "" {
		// Use current workspace
		for _, ws := range worktrees {
			if ws.IsCurrent {
				target = ws
				break
			}
		}
		if target == nil && len(worktrees) > 0 {
			target = worktrees[0]
		}
	} else {
		// Match by branch name or path suffix
		for _, ws := range worktrees {
			if ws.Branch == targetIdentifier ||
				ws.Path == targetIdentifier ||
				strings.HasSuffix(ws.Path, "/"+targetIdentifier) ||
				strings.HasSuffix(ws.Path, "\\"+targetIdentifier) {
				target = ws
				break
			}
		}
	}

	if target == nil {
		return nil, nil, fmt.Errorf("workspace %q not found", targetIdentifier)
	}

	// Fetch staged and unstaged diffs for target
	stagedDiff, _ := a.gitRunner.DiffCached(ctx, target.Path)
	unstagedDiff, _ := a.gitRunner.Diff(ctx, target.Path)

	if target.Status != nil {
		target.Status.StagedDiff = stagedDiff
		target.Status.UnstagedDiff = unstagedDiff
	}

	return target, target.Status, nil
}
