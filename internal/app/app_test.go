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

func TestApp_SpawnWorkspace(t *testing.T) {
	mainDir, _ := setupTestGitRepo(t)

	// Add .env and .env.local to mainDir
	_ = os.WriteFile(filepath.Join(mainDir, ".env"), []byte("FOO=bar\n"), 0644)
	_ = os.WriteFile(filepath.Join(mainDir, ".env.local"), []byte("LOCAL=secret\n"), 0600)
	_ = os.WriteFile(filepath.Join(mainDir, "extra.conf"), []byte("EXTRA=1\n"), 0644)

	runner, err := git.NewNativeRunner()
	if err != nil {
		t.Fatalf("failed to initialize NativeRunner: %v", err)
	}

	application := app.NewApp(runner)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Successful spawn
	res, err := application.SpawnWorkspace(ctx, mainDir, app.SpawnOptions{
		BranchName: "feat/spawn-test",
		BaseBranch: "main",
		CopyFiles:  []string{"extra.conf"},
	})
	if err != nil {
		t.Fatalf("SpawnWorkspace failed: %v", err)
	}

	if res.Workspace.Branch != "feat/spawn-test" {
		t.Errorf("expected branch 'feat/spawn-test', got %q", res.Workspace.Branch)
	}

	// Verify that destination folder was created and has files
	if _, err := os.Stat(res.Workspace.Path); err != nil {
		t.Errorf("destination workspace path does not exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(res.Workspace.Path, ".env")); err != nil {
		t.Errorf(".env was not copied to workspace")
	}
	if _, err := os.Stat(filepath.Join(res.Workspace.Path, ".env.local")); err != nil {
		t.Errorf(".env.local was not copied to workspace")
	}
	if _, err := os.Stat(filepath.Join(res.Workspace.Path, "extra.conf")); err != nil {
		t.Errorf("extra.conf was not copied to workspace")
	}

	// 2. Collision: Try to spawn the exact same branch again without --checkout
	_, err = application.SpawnWorkspace(ctx, mainDir, app.SpawnOptions{
		BranchName: "feat/spawn-test",
		BaseBranch: "main",
	})
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected collision error for existing branch, got: %v", err)
	}

	// 3. Reuse branch with CheckoutExisting:
	// First, create a branch that is not currently checked out anywhere
	runCmd(t, mainDir, "git", "branch", "feat/unattached", "main")

	customDir := filepath.Join(t.TempDir(), "wt-checkout-reuse")
	resCheckout, err := application.SpawnWorkspace(ctx, mainDir, app.SpawnOptions{
		BranchName:       "feat/unattached",
		TargetDir:        customDir,
		CheckoutExisting: true,
	})
	if err != nil {
		t.Fatalf("unexpected checkout error: %v", err)
	}
	if resCheckout.Workspace.Path != customDir {
		t.Errorf("expected customDir %q, got %q", customDir, resCheckout.Workspace.Path)
	}
	if resCheckout.Workspace.Branch != "feat/unattached" {
		t.Errorf("expected branch 'feat/unattached', got %q", resCheckout.Workspace.Branch)
	}
}

func TestApp_RemoveWorkspace_Safety(t *testing.T) {
	mainDir, secondDir := setupTestGitRepo(t)

	runner, err := git.NewNativeRunner()
	if err != nil {
		t.Fatalf("failed to initialize NativeRunner: %v", err)
	}

	application := app.NewApp(runner)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Guard against removing main workspace
	_, err = application.RemoveWorkspace(ctx, mainDir, "main", true)
	if err == nil || !strings.Contains(err.Error(), "main working tree") {
		t.Errorf("expected error removing main workspace, got: %v", err)
	}

	// 2. Guard against removing active workspace (when called from secondDir)
	_, err = application.RemoveWorkspace(ctx, secondDir, "feat/test", true)
	if err == nil || !strings.Contains(err.Error(), "active workspace") {
		t.Errorf("expected error removing active workspace, got: %v", err)
	}

	// 3. Guard against removing secondary workspace with dirty changes (secondDir has dirty.txt)
	// From mainDir, attempt to remove secondDir without force
	_, err = application.RemoveWorkspace(ctx, mainDir, "feat/test", false)
	if err == nil || !strings.Contains(err.Error(), "safety check blocked removal") {
		t.Errorf("expected safety blocked error for dirty workspace, got: %v", err)
	}

	// 4. Guard against removing secondary workspace with ignored files
	// Clean the untracked dirty.txt, but add an ignored file .env
	_ = os.Remove(filepath.Join(secondDir, "dirty.txt"))
	_ = os.WriteFile(filepath.Join(secondDir, ".gitignore"), []byte(".env*\n"), 0644)
	runCmd(t, secondDir, "git", "add", ".gitignore")
	runCmd(t, secondDir, "git", "commit", "-m", "add gitignore")
	_ = os.WriteFile(filepath.Join(secondDir, ".env"), []byte("SECRET=123\n"), 0600)

	_, err = application.RemoveWorkspace(ctx, mainDir, "feat/test", false)
	if err == nil || !strings.Contains(err.Error(), "safety check blocked removal") {
		t.Errorf("expected safety blocked error for ignored file, got: %v", err)
	}

	// 5. Successful removal with --force
	report, err := application.RemoveWorkspace(ctx, mainDir, "feat/test", true)
	if err != nil {
		t.Fatalf("failed to remove workspace with force: %v", err)
	}
	if report == nil {
		t.Fatalf("expected safety report")
	}

	// Verify that destination directory no longer exists
	if _, err := os.Stat(secondDir); !os.IsNotExist(err) {
		t.Errorf("expected secondDir to be removed, but still exists on disk")
	}

	// Verify that the branch 'feat/test' is PRESERVED in Git
	exists, err := runner.BranchExists(ctx, mainDir, "feat/test")
	if err != nil || !exists {
		t.Errorf("expected branch 'feat/test' to remain preserved in Git history, exists=%v, err=%v", exists, err)
	}
}
