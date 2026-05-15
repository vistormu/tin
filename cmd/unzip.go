package cmd

import (
	"github.com/spf13/cobra"

	"tin/internal/archive"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewUnzipCommand() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "unzip <archive.zip>",
		Short: "Extract a zip archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := archive.Unzip(archive.UnzipOptions{
				Input:     args[0],
				OutputDir: outputDir,
			})
			if err != nil {
				return err
			}

			ui.PrintBox(cmd.OutOrStdout(), "Unzipped", []ui.Row{
				{Label: "Files", Value: plural(result.Files, "file")},
				{Label: "Bytes", Value: size.Format(result.Bytes)},
				{Label: "Output", Value: result.OutputDir},
			})
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory")
	return cmd
}
