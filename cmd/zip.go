package cmd

import (
	"github.com/spf13/cobra"

	"tin/internal/archive"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewZipCommand() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "zip <path>...",
		Short: "Create a zip archive",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := archive.Zip(archive.ZipOptions{
				Inputs: args,
				Output: output,
			})
			if err != nil {
				return err
			}

			ui.PrintBox(cmd.OutOrStdout(), "Zipped", []ui.Row{
				{Label: "Files", Value: plural(result.Files, "file")},
				{Label: "Bytes", Value: size.Format(result.Bytes)},
				{Label: "Output", Value: result.Output},
			})
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output archive path")
	return cmd
}
