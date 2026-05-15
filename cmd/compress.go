package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tin/internal/compress"
	"tin/internal/shared/cobrahelp"
	"tin/internal/shared/helptext"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewCompressCommand() *cobra.Command {
	var quality string
	var output string

	cmd := &cobra.Command{
		Use:   "compress <file>",
		Short: "Compress a PDF or image file",
		Long:  "Compress a supported PDF or image file.",
		Example: `  tin compress report.pdf
  tin compress report.pdf -q screen
  tin compress photo.jpg -q 76 -o photo-small.jpg
  tin compress qualities`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := compress.Compress(cmd.Context(), compress.Options{
				Input:   args[0],
				Output:  output,
				Quality: quality,
			})
			if err != nil {
				return err
			}

			ui.PrintCompressResult(
				cmd.OutOrStdout(),
				size.Format(result.BeforeSize),
				size.Format(result.AfterSize),
				result.ReductionPercent(),
				result.Output,
			)
			return nil
		},
	}

	cobrahelp.SetQualityHelp(cmd, compress.QualityHelpSections())
	cmd.Flags().StringVarP(&quality, "quality", "q", "", "PDF preset or image quality; run 'tin compress qualities' for values")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output path")
	cmd.AddCommand(newCompressQualitiesCommand())

	return cmd
}

func newCompressQualitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "qualities",
		Aliases: []string{"quality"},
		Short:   "Show supported compression quality values",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Quality Values:")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), helptext.RenderSections(compress.QualityHelpSections()))
			return nil
		},
	}
}
