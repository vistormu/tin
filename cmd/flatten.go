package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"tin/internal/flatten"
	"tin/internal/shared/confirm"
	"tin/internal/shared/ui"
)

func NewFlattenCommand() *cobra.Command {
	var dryRun bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "flatten <dir>",
		Short: "Move nested files to a directory root",
		Long:  "Flatten a directory by moving all nested files into the directory root and removing empty subdirectories.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plan, err := flatten.BuildPlan(args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			printFlattenPlan(out, plan)
			if dryRun {
				ui.PrintNote(out, "Dry run: no files were moved")
				return nil
			}
			if len(plan.Moves) == 0 && len(plan.Dirs) == 0 {
				ui.PrintNote(out, "Nothing to flatten")
				return nil
			}

			ok, err := confirm.Ask(yes, fmt.Sprintf("Move %d file(s) and remove %d directories in %s?", len(plan.Moves), len(plan.Dirs), plan.Root), cmd.InOrStdin(), out)
			if err != nil {
				return err
			}
			if !ok {
				ui.PrintWarning(out, "Aborted")
				return nil
			}

			result, err := flatten.Apply(plan)
			if err != nil {
				return err
			}
			ui.PrintBox(out, "Flattened", []ui.Row{
				{Label: "Moved", Value: plural(result.MovedFiles, "file")},
				{Label: "Removed", Value: plural(result.RemovedDirs, "directory")},
				{Label: "Root", Value: plan.Root},
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print planned moves without changing files")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")

	return cmd
}

func printFlattenPlan(out io.Writer, plan flatten.Plan) {
	ui.PrintBox(out, "Flatten Plan", []ui.Row{
		{Label: "Root", Value: plan.Root},
		{Label: "Moves", Value: plural(len(plan.Moves), "file")},
		{Label: "Dirs", Value: plural(len(plan.Dirs), "directory")},
	})
	for _, move := range plan.Moves {
		fmt.Fprintf(out, "  %s -> %s\n", move.Source, move.Destination)
	}
	for _, dir := range plan.Dirs {
		fmt.Fprintf(out, "  remove %s\n", dir)
	}
}
