package cmd

import (
	"github.com/spf13/cobra"

	"tin/internal/shared/fsx"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewSizeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "size <path|glob>...",
		Short: "Show the total size of a file or directory",
		Long:  "Show a human-readable size for one or more files or directories. Accepts paths and shell-style glob patterns.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			for _, path := range paths {
				total, err := size.Path(path)
				if err != nil {
					return err
				}
				ui.PrintSize(out, size.Format(total), path)
			}
			return nil
		},
	}
}
