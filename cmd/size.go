package cmd

import (
	"github.com/spf13/cobra"

	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewSizeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "size <path>",
		Short: "Show the total size of a file or directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			total, err := size.Path(args[0])
			if err != nil {
				return err
			}

			ui.PrintSize(cmd.OutOrStdout(), size.Format(total), args[0])
			return nil
		},
	}
}
