package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"tin/internal/compress"
	"tin/internal/shared/cobrahelp"
	"tin/internal/shared/fsx"
	"tin/internal/shared/helptext"
	"tin/internal/shared/size"
	"tin/internal/shared/ui"
)

func NewCompressCommand() *cobra.Command {
	var intensity string
	var output string
	var inPlace bool

	cmd := &cobra.Command{
		Use:   "compress <file|glob>...",
		Short: "Compress a PDF, image, or video file",
		Long:  "Compress one or more supported PDF, image, or video files. Accepts file paths and shell-style glob patterns.",
		Example: `  tin compress report.pdf
  tin compress report.pdf -i high
  tin compress photo.jpg -i medium -o photo-small.jpg
  tin compress '*.pdf' -i high
  tin compress photo.jpg --in-place
  tin compress intensities`,
		Args: cobra.MinimumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return nil, cobra.ShellCompDirectiveDefault
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			inputs, err := fsx.ExpandPaths(args)
			if err != nil {
				return err
			}
			if len(inputs) > 1 && output != "" {
				return fmt.Errorf("--output is only allowed with a single input")
			}
			if inPlace && output != "" {
				return fmt.Errorf("--in-place cannot be combined with --output")
			}

			out := cmd.OutOrStdout()
			for _, input := range inputs {
				result, err := compress.Compress(cmd.Context(), compress.Options{
					Input:     input,
					Output:    output,
					Intensity: intensity,
					InPlace:   inPlace,
				})
				if err != nil {
					return err
				}

				ui.PrintCompressResult(
					out,
					size.Format(result.BeforeSize),
					size.Format(result.AfterSize),
					result.ReductionPercent(),
					result.Output,
				)
			}
			return nil
		},
	}

	cobrahelp.SetQualityHelp(cmd, compress.IntensityHelpSections())
	cmd.Flags().StringVarP(&intensity, "intensity", "i", "", "compression intensity: low, medium (default), high, or veryhigh")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output path (single input only)")
	cmd.Flags().BoolVar(&inPlace, "in-place", false, "replace each input file with its compressed output")
	cmd.AddCommand(newCompressIntensitiesCommand())

	return cmd
}

func newCompressIntensitiesCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "intensities",
		Aliases: []string{"intensity"},
		Short:   "Show supported compression intensity values",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Intensity Values:")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), helptext.RenderSections(compress.IntensityHelpSections()))
			return nil
		},
	}
}
