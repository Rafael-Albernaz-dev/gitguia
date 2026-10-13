package app_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Rafael-Albernaz-dev/gitguia/internal/app"
	"github.com/Rafael-Albernaz-dev/gitguia/internal/git"
)

func setupTestGitRepo(t *testing.T) (string, string) {
	t.Helper()

	mainDir := t.TempDir()
	runCmd(t, mainDir, "git", "init", "-b", "main")
	runCmd(t, mainDir, "git", "config", "user.name", "Test User")
	runCmd(t, mainDir, "git", "config", "user.email", "test@example.com")

	// Commit initial file
	dummyFile := filepath.Join(mainDir, "README.md")
	if err := os.WriteFile(dummyFile, []byte("# Test Repo\n"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}
	runCmd(t, mainDir, "git", "add", "README.md")
	runCmd(t, mainDir, "git", "commit", "-m", "initial commit")

	// Create second worktree
	secondDir := filepath.Join(t.TempDir(), "wt-feature")
	runCmd(t, mainDir, "git", "worktree", "add", "-b", "feat/test", secondDir, "main")

	// Add dirty uncommitted file in second worktree
	dirtyFile := filepath.Join(secondDir, "dirty.txt")
	if err := os.WriteFile(dirtyFile, []byte("changes\n"), 0644); err != nil {
		t.Fatalf("failed to write dirty file: %v", err)
	}

	return mainDir, secondDir
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed [%s %v]: %v\nOutput: %s", name, args, err, string(out))
	}
}

func TestApp_ResolveProject(t *testing.T) {
	runner, err := git.NewNativeRunner()
	if err != nil {
		t.Fatalf("failed to initialize NativeRunner: %v", err)
	}

	application := app.NewApp(runner)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	proj, err := application.ResolveProject(ctx, ".")
	if err != nil {
		t.Fatalf("expected ResolveProject to succeed in git repository, got: %v", err)
	}

	if proj.RootPath == "" {
		t.Errorf("expected non-empty RootPath")
	}

	if !strings.Contains(proj.GitDir, ".git") {
		t.Errorf("expected GitDir to contain .git, got: %q", proj.GitDir)
	}
}

func TestApp_ListWorktrees_And_Status(t *testing.T) {
	mainDir, secondDir := setupTestGitRepo(t)

	runner, err := git.NewNativeRunner()
	if err != nil {
		t.Fatalf("failed to initialize NativeRunner: %v", err)
	}

	application := app.NewApp(runner)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Test ListWorktrees from mainDir
	worktrees, err := application.ListWorktrees(ctx, mainDir)
	if err != nil {
		t.Fatalf("ListWorktrees failed: %v", err)
	}

	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}

	// Main worktree checks
	if !worktrees[0].IsMain {
		t.Errorf("expected worktree[0] to be main")
	}
	if !worktrees[0].IsCurrent {
		t.Errorf("expected worktree[0] to be marked as current when called from mainDir")
	}
	if worktrees[0].Status == nil || !worktrees[0].Status.IsClean() {
		t.Errorf("expected worktree[0] to be clean")
	}

	// Second worktree checks
	if worktrees[1].IsMain {
		t.Errorf("expected worktree[1] not to be main")
	}
	if worktrees[1].Branch != "feat/test" {
		t.Errorf("expected branch 'feat/test', got %q", worktrees[1].Branch)
	}
	if worktrees[1].Status == nil || worktrees[1].Status.UntrackedCount() != 1 {
		t.Errorf("expected 1 untracked file in second worktree, got %+v", worktrees[1].Status)
	}

	// 2. Test GetWorkspaceStatus for secondDir
	ws, cs, err := application.GetWorkspaceStatus(ctx, mainDir, "feat/test")
	if err != nil {
		t.Fatalf("GetWorkspaceStatus failed: %v", err)
	}

	if ws.Branch != "feat/test" {
		t.Errorf("expected branch 'feat/test', got %q", ws.Branch)
	}
	if cs.UntrackedCount() != 1 {
		t.Errorf("expected 1 untracked file, got %d", cs.UntrackedCount())
	}
	if cs.Files[0].Path != "dirty.txt" {
		t.Errorf("expected dirty.txt, got %q", cs.Files[0].Path)
	}

	_ = secondDir
}
