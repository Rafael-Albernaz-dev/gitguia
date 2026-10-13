package domain

import "time"

// Project represents the Git repository identity and its resolved settings.
type Project struct {
	ID        string `json:"id"`        // Unique project ID (typically canonical git-common-dir path)
	RootPath  string `json:"root_path"` // Path to the main repository
	GitDir    string `json:"git_dir"`   // Resolved git common dir
	DefaultBaseBranch string `json:"default_base_branch"`
}

// Workspace represents an isolated Git worktree or the main working tree.
type Workspace struct {
	Path       string    `json:"path"`
	Branch     string    `json:"branch"`
	IsMain     bool      `json:"is_main"`
	IsLocked   bool      `json:"is_locked"`
	LockReason string    `json:"lock_reason,omitempty"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}

// FileStatus represents the status of a specific file in Git.
type FileStatus struct {
	Path     string `json:"path"`
	Staged   string `json:"staged"`   // 'M', 'A', 'D', 'R', ' '
	Unstaged string `json:"unstaged"` // 'M', 'D', '?', ' '
}

// ChangeSet holds current working tree and index changes.
type ChangeSet struct {
	WorkspacePath string       `json:"workspace_path"`
	Branch        string       `json:"branch"`
	Files         []FileStatus `json:"files"`
	StagedDiff    string       `json:"staged_diff"`
	UnstagedDiff  string       `json:"unstaged_diff"`
}

// IsClean returns true if there are no staged or unstaged modifications.
func (c *ChangeSet) IsClean() bool {
	return len(c.Files) == 0
}

// SuggestionEngineType identifies the mechanism that generated the proposal.
type SuggestionEngineType string

const (
	EngineDeterministicRules SuggestionEngineType = "rules"
	EngineLocalModel         SuggestionEngineType = "local_model"
	EngineManualInput        SuggestionEngineType = "manual"
)

// CommitProposal represents a suggested commit message and its metadata.
type CommitProposal struct {
	Subject       string               `json:"subject"`
	Body          string               `json:"body,omitempty"`
	Engine        SuggestionEngineType `json:"engine"`
	ModelName     string               `json:"model_name,omitempty"`
	GeneratedAt   time.Time            `json:"generated_at"`
	IsUserEdited  bool                 `json:"is_user_edited"`
}

// Message returns the formatted commit message.
func (p *CommitProposal) Message() string {
	if p.Body == "" {
		return p.Subject
	}
	return p.Subject + "\n\n" + p.Body
}

// SafetyViolationType enumerates reasons why an operation might be blocked.
type SafetyViolationType string

const (
	ViolationDirtyWorkingTree SafetyViolationType = "dirty_working_tree"
	ViolationUntrackedFiles   SafetyViolationType = "untracked_files"
	ViolationIgnoredFiles     SafetyViolationType = "ignored_sensitive_files"
	ViolationMainWorkspace    SafetyViolationType = "cannot_remove_main_workspace"
	ViolationBranchConflict   SafetyViolationType = "branch_conflict"
	ViolationPathConflict     SafetyViolationType = "path_conflict"
)

// SafetyViolation details a specific blocked condition.
type SafetyViolation struct {
	Type        SafetyViolationType `json:"type"`
	Description string              `json:"description"`
	ItemPath    string              `json:"item_path,omitempty"`
}

// SafetyReport determines whether an operation is allowed or blocked.
type SafetyReport struct {
	IsSafe     bool              `json:"is_safe"`
	Violations []SafetyViolation `json:"violations"`
}
