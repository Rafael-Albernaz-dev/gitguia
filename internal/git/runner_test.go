package git_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Rafael-Albernaz-dev/gitguia/internal/git"
)

func TestNativeRunner_Version(t *testing.T) {
	runner, err := git.NewNativeRunner()
	if err != nil {
		t.Fatalf("failed to initialize NativeRunner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ver, err := runner.Version(ctx)
	if err != nil {
		t.Fatalf("expected git --version to succeed, got error: %v", err)
	}

	if !strings.HasPrefix(ver, "git version") {
		t.Errorf("expected version output to start with 'git version', got %q", ver)
	}
}
