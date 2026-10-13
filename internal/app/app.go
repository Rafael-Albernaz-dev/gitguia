package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/Rafael-Albernaz-dev/gitguia/internal/domain"
	"github.com/Rafael-Albernaz-dev/gitguia/internal/git"
)

// Service defines the core use case contracts available to both CLI and TUI.
type Service interface {
	GetGitVersion(ctx context.Context) (string, error)
	ResolveProject(ctx context.Context, dir string) (*domain.Project, error)
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
	// git rev-parse --path-format=absolute --git-common-dir
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
