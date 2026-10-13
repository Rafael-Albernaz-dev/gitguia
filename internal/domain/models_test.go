package domain_test

import (
	"testing"
	"time"

	"github.com/Rafael-Albernaz-dev/gitguia/internal/domain"
)

func TestChangeSet_IsClean(t *testing.T) {
	csEmpty := domain.ChangeSet{
		WorkspacePath: "/path/to/ws",
		Files:         []domain.FileStatus{},
	}
	if !csEmpty.IsClean() {
		t.Errorf("expected empty changeset to be clean")
	}

	csDirty := domain.ChangeSet{
		WorkspacePath: "/path/to/ws",
		Files: []domain.FileStatus{
			{Path: "main.go", Unstaged: "M"},
		},
	}
	if csDirty.IsClean() {
		t.Errorf("expected changeset with files not to be clean")
	}
}

func TestCommitProposal_Message(t *testing.T) {
	propSubjectOnly := domain.CommitProposal{
		Subject:     "feat: add initial foundation",
		Engine:      domain.EngineDeterministicRules,
		GeneratedAt: time.Now(),
	}
	if got := propSubjectOnly.Message(); got != "feat: add initial foundation" {
		t.Errorf("expected %q, got %q", "feat: add initial foundation", got)
	}

	propWithBody := domain.CommitProposal{
		Subject: "fix: prevent data loss",
		Body:    "Inspect ignored files before teardown.",
		Engine:  domain.EngineLocalModel,
	}
	expected := "fix: prevent data loss\n\nInspect ignored files before teardown."
	if got := propWithBody.Message(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
