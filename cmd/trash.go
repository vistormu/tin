package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tin/internal/shared/confirm"
	"tin/internal/shared/ui"
	"tin/internal/trash"
)

func NewTrashCommand() *cobra.Command {
	var dryRun bool
	var yes bool

	cmd := &cobra.Command{
		Use:   "trash <file|glob>...",
		Short: "Move files to the system trash",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targets, err := trash.ExpandTargets(args)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			ui.PrintTargets(out, targets)

			if dryRun {
				ui.PrintNote(out, "dry run: no files were moved")
				return nil
			}

			ok, err := confirm.Ask(yes, fmt.Sprintf("move %d item(s) to trash?", len(targets)), cmd.InOrStdin(), out)
			if err != nil {
				return err
			}
			if !ok {
				ui.PrintWarning(out, "aborted")
				return nil
			}

			mover, err := trash.NewMover()
			if err != nil {
				return err
			}
			for _, target := range targets {
				if err := mover.Move(target); err != nil {
					return fmt.Errorf("trash %q: %w", target, err)
				}
			}

			ui.PrintSuccess(out, fmt.Sprintf("trashed %d item(s)", len(targets)))
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print targets without trashing them")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")

	return cmd
}
