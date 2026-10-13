package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Rafael-Albernaz-dev/gitguia/internal/app"
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
			fmt.Println(titleStyle.Render("Worktree Discovery") + " (Scheduled for slice GGUIA-002)")
			fmt.Println(subtleStyle.Render("Parsing worktrees with 'git worktree list --porcelain -z'..."))
			return nil
		},
	}
}

func newSpawnCmd(appCore *app.App) *cobra.Command {
	var baseBranch string

	cmd := &cobra.Command{
		Use:   "spawn <branch-name>",
		Short: "Create an isolated Git worktree and copy declared local files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			branch := args[0]
			fmt.Println(titleStyle.Render("Spawn Workspace") + " (Scheduled for slice GGUIA-003)")
			fmt.Printf("Provisioning worktree for branch: %s (base: %s)\n", branch, baseBranch)
			return nil
		},
	}

	cmd.Flags().StringVarP(&baseBranch, "base", "b", "main", "Base branch to branch off from")
	return cmd
}

func newStatusCmd(appCore *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display workspace status, changes, and staged diffs",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(titleStyle.Render("Workspace Status") + " (Scheduled for slice GGUIA-002)")
			return nil
		},
	}
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
