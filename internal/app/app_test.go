package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Rafael-Albernaz-dev/gitguia/internal/app"
	"github.com/Rafael-Albernaz-dev/gitguia/internal/git"
)

func TestApp_ResolveProject(t *testing.T) {
	runner, err := git.NewNativeRunner()
	if err != nil {
		t.Fatalf("failed to initialize NativeRunner: %v", err)
	}

	application := app.NewApp(runner)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Resolving current project
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
