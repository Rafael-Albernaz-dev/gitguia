package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Rafael-Albernaz-dev/gitguia/internal/app"
	"github.com/Rafael-Albernaz-dev/gitguia/internal/domain"
	"github.com/Rafael-Albernaz-dev/gitguia/internal/git"
	"github.com/Rafael-Albernaz-dev/gitguia/internal/version"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	subtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575"))

	warnStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFB86C"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555"))

	branchStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8BE9FD"))

	activeMarkerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#50FA7B"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#BD93F9")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color("#44475A"))
)

// NewRootCmd initializes the root Cobra command and its subcommands.
func NewRootCmd() *cobra.Command {
	var showVersion bool

	rootCmd := &cobra.Command{
		Use:   "gitguia",
		Short: "A local-first companion for safe Git worktree workflows and guided commits",
		Long: titleStyle.Render("gitguia") + " — " +
			subtleStyle.Render("Opinionated, local-first CLI and TUI companion in Go for humans and AI agents.\n") +
			"Manage Git Worktrees, inspect staged diffs, and create safe commits without friction.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Printf("gitguia version %s (commit: %s, built: %s)\n",
					version.Version, version.GitCommit, version.BuildDate)
				return nil
			}
			return cmd.Help()
		},
	}

	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show gitguia version")

	// Wire application core
	gitRunner, err := git.NewNativeRunner()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	appCore := app.NewApp(gitRunner)

	// Add subcommands
	rootCmd.AddCommand(newDoctorCmd(appCore))
	rootCmd.AddCommand(newListCmd(appCore))
	rootCmd.AddCommand(newSpawnCmd(appCore))
	rootCmd.AddCommand(newStatusCmd(appCore))
	rootCmd.AddCommand(newCommitCmd(appCore))
	rootCmd.AddCommand(newPushCmd(appCore))
	rootCmd.AddCommand(newRmCmd(appCore))

	return rootCmd
}

// Execute runs the root command.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newDoctorCmd(appCore *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Verify environment health, native Git installation, and current workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			gitVer, err := appCore.GetGitVersion(ctx)
			if err != nil {
				return fmt.Errorf("git check failed: %w", err)
			}
			fmt.Println(successStyle.Render("✓") + " Native Git: " + gitVer)

			proj, err := appCore.ResolveProject(ctx, ".")
			if err != nil {
				fmt.Println(subtleStyle.Render("ℹ Current directory is not a Git repository (or in a bare tree)."))
			} else {
				fmt.Println(successStyle.Render("✓") + " Git Common Dir: " + proj.GitDir)
				fmt.Println(successStyle.Render("✓") + " Git Root: " + proj.RootPath)
			}

			fmt.Println(successStyle.Render("✓") + " gitguia engine initialized successfully.")
			return nil
		},
	}
}

func newListCmd(appCore *app.App) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List active worktrees and their Git statuses",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			worktrees, err := appCore.ListWorktrees(ctx, ".")
			if err != nil {
				return err
			}

			if len(worktrees) == 0 {
				fmt.Println(subtleStyle.Render("No worktrees found."))
				return nil
			}

			cwd, _ := os.Getwd()
			proj, _ := appCore.ResolveProject(ctx, ".")

			fmt.Println()
			fmt.Printf("%s  %-24s %-32s %-10s %s\n",
				headerStyle.Render("  "),
				headerStyle.Render("BRANCH"),
				headerStyle.Render("PATH"),
				headerStyle.Render("COMMIT"),
				headerStyle.Render("STATUS"),
			)

			for _, ws := range worktrees {
				marker := "  "
				if ws.IsCurrent {
					marker = activeMarkerStyle.Render("● ")
				} else if ws.IsMain {
					marker = subtleStyle.Render("○ ")
				}

				branchName := ws.Branch
				if ws.IsMain {
					branchName = branchStyle.Render(ws.Branch) + subtleStyle.Render(" (main)")
				} else {
					branchName = branchStyle.Render(ws.Branch)
				}

				// Relative display path
				displayPath := ws.Path
				if proj != nil {
					if ws.Path == proj.RootPath {
						displayPath = "./"
					} else if rel, err := filepath.Rel(filepath.Dir(proj.RootPath), ws.Path); err == nil && !strings.HasPrefix(rel, "..") {
						displayPath = "../" + rel
					} else if rel, err := filepath.Rel(cwd, ws.Path); err == nil {
						displayPath = rel
					}
				}

				shortCommit := ws.CommitHash
				if len(shortCommit) > 7 {
					shortCommit = shortCommit[:7]
				}

				// Status badges
				var statusBadge string
				if ws.IsLocked {
					statusBadge = errorStyle.Render("🔒 locked")
					if ws.LockReason != "" {
						statusBadge += subtleStyle.Render(" (" + ws.LockReason + ")")
					}
				} else if ws.Status != nil {
					if ws.Status.IsClean() {
						statusBadge = successStyle.Render("✓ clean")
					} else {
						var parts []string
						if st := ws.Status.StagedCount(); st > 0 {
							parts = append(parts, successStyle.Render(fmt.Sprintf("+%d", st)))
						}
						if unst := ws.Status.UnstagedCount(); unst > 0 {
							parts = append(parts, warnStyle.Render(fmt.Sprintf("~%d", unst)))
						}
						if untr := ws.Status.UntrackedCount(); untr > 0 {
							parts = append(parts, subtleStyle.Render(fmt.Sprintf("?%d", untr)))
						}
						statusBadge = warnStyle.Render("dirty") + " (" + strings.Join(parts, " ") + ")"
					}

					if ws.Status.Ahead > 0 || ws.Status.Behind > 0 {
						syncPart := ""
						if ws.Status.Ahead > 0 {
							syncPart += fmt.Sprintf("↑%d", ws.Status.Ahead)
						}
						if ws.Status.Behind > 0 {
							syncPart += fmt.Sprintf("↓%d", ws.Status.Behind)
						}
						statusBadge += " " + subtleStyle.Render(syncPart)
					}
				} else {
					statusBadge = subtleStyle.Render("unknown")
				}

				fmt.Printf("%s %-24s %-32s %-10s %s\n",
					marker,
					branchName,
					subtleStyle.Render(displayPath),
					subtleStyle.Render(shortCommit),
					statusBadge,
				)
			}
			fmt.Println()
			return nil
		},
	}
}

func newStatusCmd(appCore *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "status [workspace]",
		Short: "Display workspace status, changes, and staged diffs",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			target := ""
			if len(args) > 0 {
				target = args[0]
			}

			ws, cs, err := appCore.GetWorkspaceStatus(ctx, ".", target)
			if err != nil {
				return err
			}

			fmt.Println()
			header := fmt.Sprintf("Workspace: %s  |  Branch: %s",
				titleStyle.Render(ws.Path),
				branchStyle.Render(ws.Branch),
			)
			if ws.IsMain {
				header += subtleStyle.Render(" [main tree]")
			}
			fmt.Println(header)

			if len(ws.CommitHash) > 7 {
				fmt.Printf("HEAD: %s", subtleStyle.Render(ws.CommitHash[:7]))
				if cs.Ahead > 0 || cs.Behind > 0 {
					fmt.Printf(" (ahead %d, behind %d)", cs.Ahead, cs.Behind)
				}
				fmt.Println()
			}
			fmt.Println()

			if cs.IsClean() {
				fmt.Println(successStyle.Render("✓ Working tree clean (no staged or unstaged changes)"))
				fmt.Println()
				return nil
			}

			// Group files into Staged, Unstaged, and Untracked
			var stagedFiles []domain.FileStatus
			var unstagedFiles []domain.FileStatus
			var untrackedFiles []domain.FileStatus

			for _, f := range cs.Files {
				if f.Staged == "?" || f.Unstaged == "?" {
					untrackedFiles = append(untrackedFiles, f)
					continue
				}
				if f.Staged != "" && f.Staged != " " {
					stagedFiles = append(stagedFiles, f)
				}
				if f.Unstaged != "" && f.Unstaged != " " {
					unstagedFiles = append(unstagedFiles, f)
				}
			}

			if len(stagedFiles) > 0 {
				fmt.Println(successStyle.Render("Changes to be committed:") + subtleStyle.Render(" (staged for next commit)"))
				for _, f := range stagedFiles {
					pathStr := f.Path
					if f.OrigPath != "" {
						pathStr = f.OrigPath + " -> " + f.Path
					}
					fmt.Printf("  %s  %s\n", successStyle.Render(f.Staged), pathStr)
				}
				fmt.Println()
			}

			if len(unstagedFiles) > 0 {
				fmt.Println(warnStyle.Render("Changes not staged for commit:") + subtleStyle.Render(" (use 'git add <file>...')"))
				for _, f := range unstagedFiles {
					fmt.Printf("  %s  %s\n", warnStyle.Render(f.Unstaged), f.Path)
				}
				fmt.Println()
			}

			if len(untrackedFiles) > 0 {
				fmt.Println(errorStyle.Render("Untracked files:") + subtleStyle.Render(" (unversioned files)"))
				for _, f := range untrackedFiles {
					fmt.Printf("  ?  %s\n", subtleStyle.Render(f.Path))
				}
				fmt.Println()
			}

			return nil
		},
	}
}

func newSpawnCmd(appCore *app.App) *cobra.Command {
	var baseBranch string
	var targetDir string
	var checkoutExisting bool
	var copyFiles []string

	cmd := &cobra.Command{
		Use:   "spawn <branch-name>",
		Short: "Create an isolated Git worktree and copy declared local files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]
			ctx := context.Background()

			res, err := appCore.SpawnWorkspace(ctx, ".", app.SpawnOptions{
				BranchName:       branch,
				BaseBranch:       baseBranch,
				TargetDir:        targetDir,
				CheckoutExisting: checkoutExisting,
				CopyFiles:        copyFiles,
			})
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Printf("%s Workspace provisioned successfully!\n", successStyle.Render("✓"))
			fmt.Println()
			fmt.Printf("  %-10s %s\n", subtleStyle.Render("Branch:"), branchStyle.Render(res.Workspace.Branch))
			fmt.Printf("  %-10s %s\n", subtleStyle.Render("Path:"), titleStyle.Render(res.Workspace.Path))
			if len(res.CopiedFiles) > 0 {
				fmt.Printf("  %-10s %s\n", subtleStyle.Render("Copied:"), successStyle.Render(strings.Join(res.CopiedFiles, ", ")))
			}
			fmt.Println()
			fmt.Println(subtleStyle.Render("To switch to your new workspace:"))
			fmt.Printf("  cd %s\n\n", res.Workspace.Path)
			return nil
		},
	}

	cmd.Flags().StringVarP(&baseBranch, "base", "b", "main", "Base branch to branch off from")
	cmd.Flags().StringVarP(&targetDir, "dir", "d", "", "Custom destination directory for the worktree")
	cmd.Flags().BoolVarP(&checkoutExisting, "checkout", "c", false, "Checkout existing branch instead of creating new")
	cmd.Flags().StringSliceVar(&copyFiles, "copy", nil, "Extra files to copy from the main workspace (can be repeated)")
	return cmd
}

func newCommitCmd(appCore *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "Stage files and create a commit guided by local message proposals",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(titleStyle.Render("Guided Commit") + " (Scheduled for slice GGUIA-005)")
			return nil
		},
	}
}

func newPushCmd(appCore *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Explicitly push local commits to remote target",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(titleStyle.Render("Explicit Push") + " (Scheduled for slice GGUIA-007)")
			return nil
		},
	}
}

func newRmCmd(appCore *app.App) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "rm <worktree-name>",
		Short: "Safely teardown a worktree checking for uncommitted and ignored files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			fmt.Println(titleStyle.Render("Safe Teardown") + " (Scheduled for slice GGUIA-004)")
			fmt.Printf("Inspecting worktree safety for: %s (force: %v)\n", target, force)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force teardown bypassing safety checks")
	return cmd
}
