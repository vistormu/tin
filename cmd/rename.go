package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"tin/internal/rename"
	"tin/internal/shared/confirm"
	"tin/internal/shared/fsx"
	"tin/internal/shared/ui"
)

func NewRenameCommand() *cobra.Command {
	var dryRun bool
	var yes bool
	var styleFlag string

	cmd := &cobra.Command{
		Use:   "rename <file|glob>...",
		Short: "Normalize file names to a chosen case style",
		Long:  "Rename files so each basename follows a chosen case style (kebab by default). Accepts file paths and shell-style glob patterns.",
		Example: `  tin rename 'My Photo.JPG'
  tin rename '*.JPG' --dry-run
  tin rename report.PDF --style snake
  tin rename 'My File.txt' --style camel --yes`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			style, err := rename.ParseStyle(styleFlag)
			if err != nil {
				return err
			}

			paths, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}

			plan, err := rename.BuildPlan(paths, style)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			printRenamePlan(out, plan)

			if dryRun {
				ui.PrintNote(out, "dry run: no files were renamed")
				return nil
			}
			if len(plan.Moves) == 0 {
				ui.PrintNote(out, "nothing to rename")
				return nil
			}

			ok, err := confirm.Ask(yes, fmt.Sprintf("rename %d file(s)?", len(plan.Moves)), cmd.InOrStdin(), out)
			if err != nil {
				return err
			}
			if !ok {
				ui.PrintWarning(out, "aborted")
				return nil
			}

			result, err := rename.Apply(plan)
			if err != nil {
				return err
			}
			ui.PrintBox(out, "renamed", []ui.Row{
				{Label: "moved", Value: ui.Plural(result.Moved, "file")},
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned renames without changing files")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	cmd.Flags().StringVarP(&styleFlag, "style", "s", "", "case style: kebab (default), snake, camel, pascal, or screaming")

	return cmd
}

func printRenamePlan(out io.Writer, plan rename.Plan) {
	ui.PrintBox(out, "rename plan", []ui.Row{
		{Label: "renames", Value: ui.Plural(len(plan.Moves), "file")},
		{Label: "skipped", Value: ui.Plural(len(plan.NoOps), "file")},
	})
	for _, move := range plan.Moves {
		fmt.Fprintf(out, "  %s -> %s\n", move.Source, move.Destination)
	}
}
