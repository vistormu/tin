package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	githelper "tin/internal/git"
	"tin/internal/shared/confirm"
	"tin/internal/shared/ui"
)

func NewGitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "git",
		Short: "Run safer helpers for common git operations",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(
		newGitUndoCommand(),
		newGitObliterateCommand(),
	)

	return cmd
}

func newGitUndoCommand() *cobra.Command {
	var hard bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Undo the latest commit",
		Long:  "Undo the latest commit with a soft reset, leaving its changes staged. With --hard, discard the commit and all tracked index and working-tree changes. Untracked files are never removed.",
		Example: `  tin git undo
  tin git undo --hard
  tin git undo --hard --yes`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gitClient()
			if err != nil {
				return err
			}
			plan, err := client.BuildUndoPlan(cmd.Context())
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			mode := "soft (changes remain staged)"
			if hard {
				mode = "hard (tracked changes are discarded)"
			}
			ui.PrintBox(out, "git undo", []ui.Row{
				{Label: "repository", Value: plan.Root},
				{Label: "commit", Value: fmt.Sprintf("%s %s", plan.Commit.ShortHash, plan.Commit.Subject)},
				{Label: "mode", Value: mode},
			})

			if hard {
				ui.PrintWarning(out, fmt.Sprintf(
					"this will discard the commit plus %s; %s will remain",
					ui.Plural(plan.TrackedChanges, "tracked change"),
					ui.Plural(plan.UntrackedFiles, "untracked file"),
				))
				ok, err := confirm.Ask(yes, "permanently discard these tracked changes?", cmd.InOrStdin(), out)
				if err != nil {
					return err
				}
				if !ok {
					ui.PrintWarning(out, "aborted")
					return nil
				}
			}

			if err := client.Undo(cmd.Context(), plan, hard); err != nil {
				return err
			}
			ui.PrintBox(out, "commit undone", []ui.Row{
				{Label: "commit", Value: plan.Commit.ShortHash},
				{Label: "mode", Value: map[bool]string{false: "soft", true: "hard"}[hard]},
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&hard, "hard", false, "discard the commit and all tracked local changes")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation for --hard")

	return cmd
}

func newGitObliterateCommand() *cobra.Command {
	var dryRun bool
	var force bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "obliterate <path>",
		Short: "Remove a path from the entire git history",
		Long:  "Permanently remove an exact file or directory path from all local git history using git-filter-repo. The working tree must be clean. Paths are normalized relative to the repository root, and historical renames are not followed.",
		Example: `  tin git obliterate .env --dry-run
  tin git obliterate config/secrets.json
  tin git obliterate old/path/to/secret --force --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := gitClient()
			if err != nil {
				return err
			}
			plan, err := client.BuildObliteratePlan(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			rows := []ui.Row{
				{Label: "repository", Value: plan.Root},
				{Label: "path", Value: plan.Path},
				{Label: "commits", Value: ui.Plural(plan.AffectedCommits, "commit")},
			}
			if plan.Origin != "" {
				rows = append(rows, ui.Row{Label: "origin", Value: plan.Origin})
			}
			ui.PrintBox(out, "git obliterate plan", rows)
			ui.PrintWarning(out, "this rewrites commit hashes across local branches and tags")
			ui.PrintNote(out, "git-filter-repo may remove the origin remote; historical renames are not followed")

			if dryRun {
				ui.PrintNote(out, "dry run: repository history was not changed")
				return nil
			}

			ok, err := confirm.Ask(yes, fmt.Sprintf("obliterate %q from repository history?", plan.Path), cmd.InOrStdin(), out)
			if err != nil {
				return err
			}
			if !ok {
				ui.PrintWarning(out, "aborted")
				return nil
			}

			filterOutput, err := client.Obliterate(cmd.Context(), plan, force)
			if err != nil {
				return err
			}
			if filterOutput != "" {
				ui.PrintNote(out, filterOutput)
			}
			ui.PrintBox(out, "path obliterated", []ui.Row{
				{Label: "path", Value: plan.Path},
				{Label: "commits", Value: ui.Plural(plan.AffectedCommits, "commit")},
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print the rewrite plan without changing history")
	cmd.Flags().BoolVar(&force, "force", false, "override git-filter-repo's fresh-clone safety check")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")

	return cmd
}

func gitClient() (*githelper.Client, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return githelper.NewClient(workDir), nil
}
