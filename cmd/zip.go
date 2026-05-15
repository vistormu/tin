package cmd

import (
	"github.com/spf13/cobra"

	"tin/internal/archive"
	"tin/internal/shared/fsx"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewZipCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "zip <path|glob>...",
		Short: "Create a zip archive",
		Long:  "Create a zip archive from one or more files or directories. Accepts paths and shell-style glob patterns.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}

			result, err := archive.Zip(archive.ZipOptions{
				Inputs: inputs,
				Output: output,
			})
			if err != nil {
				return err
			}

			ui.PrintBox(cmd.OutOrStdout(), "Zipped", []ui.Row{
				{Label: "Files", Value: ui.Plural(result.Files, "file")},
				{Label: "Bytes", Value: size.Format(result.Bytes)},
				{Label: "Output", Value: result.Output},
			})
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output archive path")
	return cmd
}
