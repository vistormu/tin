package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"tin/internal/dedup"
	"tin/internal/shared/confirm"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
	"tin/internal/trash"
)

func NewDedupCommand() *cobra.Command {
	var dryRun bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "dedup <dir>",
		Short: "Find duplicate files and move extras to trash",
		Long:  "Detect duplicate files under a directory by content hash. The alphabetically-first copy of each group is kept; the rest are moved to the system trash. Compares files by size first and only hashes candidates that share a size.",
		Example: `  tin dedup ./photos --dry-run
  tin dedup ./downloads --yes`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			plan, err := dedup.BuildPlan(args[0])
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			printDedupPlan(out, plan)

			if dryRun {
				ui.PrintNote(out, "Dry run: no files were trashed")
				return nil
			}
			removable := plan.Removable()
			if removable == 0 {
				ui.PrintNote(out, "No duplicates found")
				return nil
			}

			ok, err := confirm.Ask(yes, fmt.Sprintf("Trash %d duplicate file(s)?", removable), cmd.InOrStdin(), out)
			if err != nil {
				return err
			}
			if !ok {
				ui.PrintWarning(out, "Aborted")
				return nil
			}

			mover, err := trash.NewMover()
			if err != nil {
				return err
			}

			result, err := dedup.Apply(plan, mover)
			if err != nil {
				return err
			}

			ui.PrintBox(out, "Deduped", []ui.Row{
				{Label: "Trashed", Value: ui.Plural(result.Trashed, "file")},
			})
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print duplicate plan without trashing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")

	return cmd
}

func printDedupPlan(out io.Writer, plan dedup.Plan) {
	ui.PrintBox(out, "Dedup Plan", []ui.Row{
		{Label: "Groups", Value: ui.Plural(len(plan.Groups), "group")},
		{Label: "Removable", Value: ui.Plural(plan.Removable(), "file")},
		{Label: "Recoverable", Value: size.Format(plan.RecoverableBytes())},
	})
	for i, g := range plan.Groups {
		fmt.Fprintf(out, "  group %d (%s)\n", i+1, size.Format(g.Size))
		fmt.Fprintf(out, "    keep   %s\n", g.Keep)
		for _, t := range g.Trash {
			fmt.Fprintf(out, "    trash  %s\n", t)
		}
	}
}
